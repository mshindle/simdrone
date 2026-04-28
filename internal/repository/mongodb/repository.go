package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/repository"
	"github.com/mshindle/simdrone/internal/telemetry"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

const (
	positionsCollection = "positions"
	telemetryCollection = "telemetry"
	alertsCollection    = "alerts"

	attrDBCollection = "db.collection"
)

type EventRollupRepository struct {
	opts                *options.ClientOptions
	dbname              string
	client              *mongo.Client
	db                  *mongo.Database
	positionsCollection *mongo.Collection
	telemetryCollection *mongo.Collection
	alertsCollection    *mongo.Collection
	tracer              trace.Tracer
}

func New(dbname string, tracer trace.Tracer, opts *options.ClientOptions) *EventRollupRepository {
	return &EventRollupRepository{opts: opts, dbname: dbname, tracer: tracer}
}

func insertOneCollection[T mongoRecord](ctx context.Context, tracer trace.Tracer, collection *mongo.Collection, record *T) error {
	colName := collection.Name()
	spanName := fmt.Sprintf("mongodb.Insert.%s", upperFirst(colName))
	ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.SetAttributes(
		attribute.String(telemetry.AttrDBSystem, "mongodb"),
		attribute.String(attrDBCollection, colName),
	)
	_, err := collection.InsertOne(ctx, record)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

// AddAlert inserts a new alert event into the alerts collection
func (r *EventRollupRepository) AddAlert(ctx context.Context, evt *event.AlertSignalled) error {
	record := convertAlertEventToRecord(evt, bson.NewObjectID())
	return insertOneCollection[mongoAlertRecord](ctx, r.tracer, r.alertsCollection, record)
}

// AddTelemetry inserts a new telemetry event into the telemetry collection
func (r *EventRollupRepository) AddTelemetry(ctx context.Context, evt *event.TelemetryUpdated) error {
	record := convertTelemetryEventToRecord(evt, bson.NewObjectID())
	return insertOneCollection[mongoTelemetryRecord](ctx, r.tracer, r.telemetryCollection, record)

}

// AddPosition inserts a new position event into the positions collection
func (r *EventRollupRepository) AddPosition(ctx context.Context, evt *event.PositionChanged) error {
	record := convertPositionEventToRecord(evt, bson.NewObjectID())
	return insertOneCollection[mongoPositionRecord](ctx, r.tracer, r.positionsCollection, record)
}

func (r *EventRollupRepository) GetAlert(ctx context.Context, droneID string) (*event.AlertSignalled, error) {
	var result mongoAlertRecord
	filter := bson.M{"drone_id": droneID}
	opts := options.FindOne().SetSort(bson.D{{Key: "received_at", Value: -1}})

	err := r.alertsCollection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound.Wrap(err)
		}
		return nil, err
	}
	return convertAlertRecordToEvent(&result), nil
}

func (r *EventRollupRepository) GetTelemetry(ctx context.Context, droneID string) (*event.TelemetryUpdated, error) {
	var result mongoTelemetryRecord
	filter := bson.M{"drone_id": droneID}
	opts := options.FindOne().SetSort(bson.D{{Key: "received_at", Value: -1}})

	err := r.telemetryCollection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound.Wrap(err)
		}
		return nil, err
	}
	return convertTelemetryRecordToEvent(&result), nil
}

func (r *EventRollupRepository) GetPosition(ctx context.Context, droneID string) (*event.PositionChanged, error) {
	var result mongoPositionRecord
	filter := bson.M{"drone_id": droneID}
	opts := options.FindOne().SetSort(bson.D{{Key: "received_at", Value: -1}})

	err := r.positionsCollection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound.Wrap(err)
		}
		return nil, err
	}
	return convertPositionRecordToEvent(&result), nil
}

func (r *EventRollupRepository) GetActiveDrones(ctx context.Context, d time.Duration) ([]string, error) {
	cutoff := time.Now().Add(-d)
	filter := bson.M{"received_at": bson.M{"$gte": cutoff}}

	drones := make(map[string]struct{})
	collections := []*mongo.Collection{
		r.positionsCollection,
		r.telemetryCollection,
		r.alertsCollection,
	}

	for _, col := range collections {
		var results []any
		err := col.Distinct(ctx, "drone_id", filter).Decode(&results)
		if err != nil {
			return nil, err
		}
		for _, id := range results {
			if droneID, ok := id.(string); ok {
				drones[droneID] = struct{}{}
			}
		}
	}

	activeDrones := make([]string, 0, len(drones))
	for droneID := range drones {
		activeDrones = append(activeDrones, droneID)
	}
	return activeDrones, nil
}

func (r *EventRollupRepository) Connect(ctx context.Context) error {
	client, err := mongo.Connect(r.opts)
	if err != nil {
		return err
	}
	// 2. Ping to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return err
	}

	db := client.Database(r.dbname)
	r.client = client
	r.db = db
	r.positionsCollection = db.Collection(positionsCollection)
	r.telemetryCollection = db.Collection(telemetryCollection)
	r.alertsCollection = db.Collection(alertsCollection)

	return nil
}

func (r *EventRollupRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

var Module = fx.Module("repository",
	fx.Provide(
		func(cfg *config.Config, tracer trace.Tracer) *EventRollupRepository {
			opts := options.Client().ApplyURI(cfg.Database.DSN)
			return New(cfg.Database.Name, tracer, opts)
		}),
	fx.Invoke(
		func(lc fx.Lifecycle, r *EventRollupRepository) {
			lc.Append(fx.Hook{
				OnStart: func(startCtx context.Context) error {
					return r.Connect(startCtx)
				},
				OnStop: func(stopCtx context.Context) error {
					return r.Close(stopCtx)
				},
			})
		},
	),
)
