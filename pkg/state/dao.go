package state

import (
	"context"
	"errors"

	"github.com/naonao2323/testgen/pkg/executor"
)

type State[A executor.ExecuteStrategy] interface {
	Run(ctx context.Context, event chan DaoEvent[A]) error
}

type daoStateMachine[A executor.ExecuteStrategy] struct {
	templateExecutor executor.Executor[executor.TemplateResult]
	treeExecutor     executor.Executor[executor.TreeResult]
	tableExecutor    executor.Executor[executor.TableResult]
	outputExecutor   executor.Executor[executor.OutputResult]
}

type DaoState = int

const (
	DaoStatePrepare DaoState = iota
	DaoStateTemplate
	DaoStateExecute
	DaoStateDone
)

type DaoEvent[A executor.ExecuteStrategy] struct {
	state  DaoState
	result A
}

func NewDaoState[A executor.ExecuteStrategy](
	templateExecutor executor.Executor[executor.TemplateResult],
	treeExecutor executor.Executor[executor.TreeResult],
	tableExecutor executor.Executor[executor.TableResult],
	outputExecutor executor.Executor[executor.OutputResult],
) State[A] {
	return &daoStateMachine[A]{
		templateExecutor: templateExecutor,
		treeExecutor:     treeExecutor,
		tableExecutor:    tableExecutor,
		outputExecutor:   outputExecutor,
	}
}

func (s *daoStateMachine[A]) Run(ctx context.Context, event chan DaoEvent[A]) error {
	spwan := func(e DaoEvent[A]) error {
		select {
		case event <- e:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	transition := func(state DaoEvent[A]) error {
		// do sometihng
		// 純粋な処理の流れを書く
		// TODO: executeの返りの値とstateの関係性がないので、考慮が必要である。
		result := new(A)
		switch state.state {
		case DaoStatePrepare:
			resp, err := s.tableExecutor.Execute()
			if err != nil {
				return err
			}
			result = any(&resp).(*A)
		case DaoStateTemplate:
			resp, err := s.templateExecutor.Execute()
			if err != nil {
				return err
			}
			result = any(&resp).(*A)
		case DaoStateExecute:
			resp, err := s.outputExecutor.Execute()
			if err != nil {
				return err
			}
			result = any(&resp).(*A)
		case DaoStateDone:
			return nil
		}
		mState := s.mutate(state.state)
		if mState == -1 {
			return errors.New("unknown state")
		}
		if err := spwan(DaoEvent[A]{mState, *result}); err != nil {
			return err
		}
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-event:
			if !ok {
				return errors.New("consume event")
			}
			state := s.trigger(e.state)
			e.state = state
			if err := transition(e); err != nil {
				return err
			}
		}
	}
}

func (s daoStateMachine[A]) trigger(event int) DaoState {
	switch event {
	case 0:
		return DaoStatePrepare
	case 1:
		return DaoStateTemplate
	case 2:
		return DaoStateExecute
	case 3:
		return DaoStateDone
	default:
		return -1
	}
}

func (s daoStateMachine[A]) mutate(state DaoState) DaoState {
	switch state {
	case DaoStatePrepare:
		return DaoStateTemplate
	case DaoStateTemplate:
		return DaoStateExecute
	case DaoStateExecute:
		return DaoStateDone
	default:
		return -1
	}
}
