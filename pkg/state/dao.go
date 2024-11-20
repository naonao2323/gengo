package state

import (
	"context"
	"errors"

	"github.com/naonao2323/testgen/pkg/common"
	"github.com/naonao2323/testgen/pkg/executor"
	"github.com/naonao2323/testgen/pkg/executor/output"
	"github.com/naonao2323/testgen/pkg/executor/table"
)

type State interface {
	Run(ctx context.Context, event chan DaoEvent) error
}

type daoStateMachine struct {
	cancel         func()
	treeExecutor   executor.TreeExecutor
	tableExecutor  table.TableExecutor
	outputExecutor output.OutputExecutor
}

type DaoState = int

const (
	DaoStatePrepare DaoState = iota
	DaoStateExecute
	DaoStateDone
)

type (
	Props struct {
		*executor.StartResult
		*table.TableResult
		*executor.TreeResult
		*output.OutputResult
	}
	DaoEvent struct {
		State   DaoState
		Request common.Request
		Props
	}
)

func NewDaoState(
	cancel func(),
	treeExecutor executor.TreeExecutor,
	tableExecutor table.TableExecutor,
	outputExecutor output.OutputExecutor,
) State {
	return &daoStateMachine{
		cancel:         cancel,
		treeExecutor:   treeExecutor,
		tableExecutor:  tableExecutor,
		outputExecutor: outputExecutor,
	}
}

func (s *daoStateMachine) Run(ctx context.Context, events chan DaoEvent) error {
	spawn := func(state DaoState, request common.Request, props Props, events chan<- DaoEvent) error {
		mState := s.mutate(state)
		if mState == -1 {
			return errors.New("unknown state")
		}
		event := DaoEvent{
			State:   mState,
			Request: request,
			Props:   props,
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case events <- event:
			return nil
		}
	}
	transition := func(state DaoEvent) error {
		switch state.State {
		case DaoStatePrepare:
			if state.StartResult == nil {
				return nil
			}
			target := *state.StartResult
			result, err := s.tableExecutor.Execute(target.Table)
			if err != nil {
				return err
			}
			if err := spawn(state.State, state.Request, Props{TableResult: &result}, events); err != nil {
				return err
			}
		case DaoStateExecute:
			if state.TableResult == nil {
				return nil
			}
			target := *state.TableResult
			result, err := s.outputExecutor.Execute(state.Request, target.Table, target.Clumns, target.Pk)
			if err != nil {
				return err
			}
			if err := spawn(state.State, state.Request, Props{OutputResult: result}, events); err != nil {
				return err
			}
		case DaoStateDone:
			s.cancel()
			return nil
		}
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-events:
			if !ok {
				return errors.New("consume event")
			}
			state := s.trigger(e.State)
			e.State = state
			if err := transition(e); err != nil {
				return err
			}
		}
	}
}

func (s daoStateMachine) trigger(event int) DaoState {
	switch event {
	case 0:
		return DaoStatePrepare
	case 1:
		return DaoStateExecute
	case 2:
		return DaoStateDone
	default:
		return -1
	}
}

func (s daoStateMachine) mutate(state DaoState) DaoState {
	switch state {
	case DaoStatePrepare:
		return DaoStateExecute
	case DaoStateExecute:
		return DaoStateDone
	default:
		return -1
	}
}
