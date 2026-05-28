package controller_admin

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/task_square"
	"github.com/gin-gonic/gin"
)

type TaskSquareAdminController struct {
	taskSquareService *task_square.TaskSquareService
}

func NewTaskSquareAdminController(taskSquareService *task_square.TaskSquareService) *TaskSquareAdminController {
	return &TaskSquareAdminController{taskSquareService: taskSquareService}
}

func (ctrl *TaskSquareAdminController) ListTasks(ctx *gin.Context) {
	req := &schema.TaskListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = true
	resp, err := ctrl.taskSquareService.ListTasks(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

func (ctrl *TaskSquareAdminController) ReviewTask(ctx *gin.Context) {
	req := &schema.TaskReviewReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.OperatorID = middleware.GetLoginUserIDFromContext(ctx)
	err := ctrl.taskSquareService.ReviewTask(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

func (ctrl *TaskSquareAdminController) AssignTask(ctx *gin.Context) {
	req := &schema.TaskAssignReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.OperatorID = middleware.GetLoginUserIDFromContext(ctx)
	err := ctrl.taskSquareService.AssignTask(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

func (ctrl *TaskSquareAdminController) ReviewSubmission(ctx *gin.Context) {
	req := &schema.TaskSubmissionReviewReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.OperatorID = middleware.GetLoginUserIDFromContext(ctx)
	err := ctrl.taskSquareService.ReviewSubmission(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

func (ctrl *TaskSquareAdminController) FeaturePost(ctx *gin.Context) {
	req := &schema.FeaturedPostCreateReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.OperatorID = middleware.GetLoginUserIDFromContext(ctx)
	err := ctrl.taskSquareService.FeaturePost(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

func (ctrl *TaskSquareAdminController) RevokeFeaturedPost(ctx *gin.Context) {
	req := &schema.FeaturedPostRevokeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.OperatorID = middleware.GetLoginUserIDFromContext(ctx)
	err := ctrl.taskSquareService.RevokeFeaturedPostReward(ctx, req.QuestionID, req.OperatorID)
	handler.HandleResponse(ctx, err, nil)
}

func (ctrl *TaskSquareAdminController) ListFeaturedPosts(ctx *gin.Context) {
	req := &schema.FeaturedPostListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := ctrl.taskSquareService.ListFeaturedPosts(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
