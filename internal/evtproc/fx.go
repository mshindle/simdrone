package evtproc

import "go.uber.org/fx"

var Module = fx.Module("processor",
	fx.Provide(
		fx.Annotate(
			NewServer,
			fx.ParamTags(`group:"processorOptions"`),
		),
	),
)

func AsOption(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"processorOptions"`),
	)
}
