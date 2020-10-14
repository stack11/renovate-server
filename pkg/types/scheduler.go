package types

type Scheduler interface {
	Schedule(args ExecutionArgs) error
}
