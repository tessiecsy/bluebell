package controller

import (
	"bluebell/logic"
	"bluebell/models"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

// 投票

func PostVoteHandler(c *gin.Context) {
	// 参数校验
	p := new(models.ParamVoteData)
	if err := c.ShouldBindJSON(p); err != nil {
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		errData := removeTopStruct(errs.Translate(trans))  // 翻译去除错误提示中的结构体标识（返回前端的时候）
		ResponseErrorWithMsg(c, CodeInvalidParam, errData) // 数据返回给前端
		return
	}
	userID, err := getCurrentUserID(c) // 获取当前登录的用户id
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 具体投票逻辑
	if err := logic.VoteForPost(userID, p); err != nil {
		zap.L().Error("logic.VoteForPost(userID, p) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)

}
