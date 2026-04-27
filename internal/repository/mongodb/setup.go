package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/mshindle/structures/set"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func EnsureIndexes(ctx context.Context, e *EventRollupRepository) error {
	collections := set.New[*mongo.Collection](3)
	collections.Add(e.alertsCollection)
	collections.Add(e.telemetryCollection)
	collections.Add(e.positionsCollection)

	builder := options.Index().SetName("idx_drone_latest")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "drone_id", Value: 1},
			{Key: "received_at", Value: -1},
		},
		Options: builder,
	}

	var errs error
	for coll := range collections.All() {
		if _, err := coll.Indexes().CreateOne(ctx, indexModel); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to create index for collection %s: %w", coll.Name(), err))
		}
	}

	return errs
}
