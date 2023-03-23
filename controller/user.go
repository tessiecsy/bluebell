package controller

import (
	"bluebell/dao/mysql"
	"bluebell/logic"
	"bluebell/models"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// 处理注册请求的函数

func SignUpHandler(c *gin.Context) {
	// 1.获取参数和参数校验
	p := new(models.ParamSignUp)
	// ShouldBind是动态获取u的类型，并找到这个结构体里的值，要想被找到值，u结构体内的值必须首字母大写
	if err := c.ShouldBindJSON(&p); err != nil {
		//请求参数有误,直接返回响应
		//ShouldBindJSON 字段类型对不对、是不是json格式
		zap.L().Error("signup with invalid param", zap.Error(err))
		// 判断err是不是validator.ValidationErrors 类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	//手动对请求参数进行详细的业务规则校验
	/*	if len(p.Username) == 0 || len(p.Password) == 0 || len(p.RePassword) == 0 || p.Password != p.RePassword {
		zap.L().Error("signup with invalid param")
		c.JSON(http.StatusOK, gin.H{
			"msg": "请求参数有误",
		})
		return
	}*/

	// 2.业务处理
	if err := logic.SignUp(p); err != nil {
		zap.L().Error("logic.signup failed", zap.Error(err))
		if errors.Is(err, mysql.ErrorUserExist) { //用户已存在的错误
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	// 3.返回响应
	ResponseSuccess(c, nil)
}

func LoginHandler(c *gin.Context) {
	// 获取请求参数与参数校验
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("login with invalid param", zap.Error(err))
		// 判断err是不是validator.ValidationErrors 类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 业务处理
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("logic.login failed", zap.String("username", p.Username), zap.Error(err))
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
		}
		ResponseError(c, CodeInvalidPassword)
		return
	}
	// 返回响应
	ResponseSuccess(c, gin.H{
		"user_id":   fmt.Sprintf("%d", user.UserID), // id值大于1<<53-1  int64类型最大值是1<<64-1
		"user_name": user.Username,
		"token":     user.Token,
	})

}

// 数据失真
// 后端：int64转换为string
// 序列化：后端数据--JSON格式的数据
// 前端：string转换为int64
// 反序列化：JSON格式的数据--go语言中的数据
// `json:"id,string"`  在序列化的时候就把int64转化为string
