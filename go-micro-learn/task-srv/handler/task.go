package handler

import (
	"context"
	"github.com/pkg/errors"
	pb "go-micro-learn/task-srv/proto/task"
	"go-micro-learn/task-srv/repository"
)

// 要实现接口，首先当然要定义一个结构体
type TaskHandler struct {
	TaskRepository repository.TaskRepository
}

func (t *TaskHandler) Create(ctx context.Context, req *pb.Task, resp *pb.EditResponse) error {
	if req.Body == "" || req.StartTime <= 0 || req.EndTime <= 0 {
		return errors.New("bad param")
	}
	if err := t.TaskRepository.InsertOne(ctx, req); err != nil {
		return err
	}
	resp.Msg = "success"
	return nil
}
func (t *TaskHandler) Delete(ctx context.Context, req *pb.Task, resp *pb.EditResponse) error {
	if req.Id == "" {
		return errors.New("bad param")
	}
	if err := t.TaskRepository.Delete(ctx, req.Id); err != nil {
		return err
	}
	resp.Msg = req.Id
	return nil
}
func (t *TaskHandler) Modify(ctx context.Context, req *pb.Task, resp *pb.EditResponse) error {
	if req.Id == "" || req.Body == "" || req.StartTime <= 0 || req.EndTime <= 0 {
		return errors.New("bad param")
	}
	if err := t.TaskRepository.Modify(ctx, req); err != nil {
		return err
	}
	resp.Msg = "success"
	return nil
}
func (t *TaskHandler) Finished(ctx context.Context, req *pb.Task, resp *pb.EditResponse) error {
	if req.Id == "" || req.IsFinished != repository.UnFinished && req.IsFinished != repository.Finished {
		return errors.New("bad param")
	}
	if err := t.TaskRepository.Finished(ctx, req); err != nil {
		return err
	}
	resp.Msg = "success"
	return nil
}
func (t *TaskHandler) Search(ctx context.Context, req *pb.SearchRequest, resp *pb.SearchResponse) error {
	count, err := t.TaskRepository.Count(ctx, req.Keyword)
	if err != nil {
		return errors.WithMessage(err, "count row number")
	}
	if req.PageCode <= 0 {
		req.PageCode = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "createTime"
	}
	if req.Order == 0 {
		req.Order = -1
	}
	if req.PageSize*(req.PageCode-1) > count {
		return errors.New("There's not that much data")
	}
	rows, err := t.TaskRepository.Search(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "search data")
	}
	*resp = pb.SearchResponse{
		PageCode: req.PageCode,
		PageSize: req.PageSize,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Rows:     rows,
	}
	return nil
}