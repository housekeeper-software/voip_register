package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"jingxi.cn/voip_register/conf"
	"jingxi.cn/voip_register/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Controller struct {
	srv        *http.Server
	serverConf *conf.ServerConfig
}

func NewController(serverConf *conf.ServerConfig) *Controller {
	return &Controller{
		srv:        nil,
		serverConf: serverConf,
	}
}

func NoResponse(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404, page not exists!",
	})
}

func (c *Controller) Run(httpAddr string) error {
	router := gin.Default()
	router.POST("/acc/register", c.registerHandlerFunc)
	router.POST("/acc/unRegister", c.unRegisterHandlerFunc)
	router.POST("/acc/add", c.addHandlerFunc)
	router.POST("/acc/delete", c.deleteHandlerFunc)
	router.POST("/acc/commit", c.commitHandlerFunc)
	router.NoRoute(NoResponse)

	c.srv = &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}
	fmt.Printf("http server listen on: %s\n", httpAddr)

	if err := c.srv.ListenAndServe(); err != nil {
		logrus.Errorf("gin ListenAndServe(%s) error: %+v", httpAddr, err)
		return err
	}
	return nil
}

func (c *Controller) Stop() {
	if c.srv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.srv.Shutdown(ctx); err != nil {
		logrus.Errorf("gin Shutdown error: %+v", err)
	}
}

func (c *Controller) registerHandlerFunc(ctx *gin.Context) {
	data := ctx.PostForm("data")
	logrus.Infof("Register User : %+v", data)

	if len(data) < 1 {
		writeResponse(ctx, http.StatusBadRequest, "Invalid request format!", "")
		return
	}
	code, msg, userId := c.addUser(data)
	if code != http.StatusOK {
		writeResponse(ctx, code, msg, userId)
		return
	}
	code, msg = c.reloadXML()
	writeResponse(ctx, code, msg, userId)
}

func (c *Controller) unRegisterHandlerFunc(ctx *gin.Context) {
	data := ctx.PostForm("data")
	logrus.Infof("UnRegister User : %+v", data)

	if len(data) < 1 {
		writeResponse(ctx, http.StatusBadRequest, "Invalid request format!", "")
		return
	}

	code, msg, userId := c.deleteUser(data)
	if code != http.StatusOK {
		writeResponse(ctx, code, msg, userId)
		return
	}
	code, msg = c.reloadXML()
	writeResponse(ctx, code, msg, userId)
}

func (c *Controller) addHandlerFunc(ctx *gin.Context) {
	data := ctx.PostForm("data")
	logrus.Infof("Add User : %+v", data)

	if len(data) < 1 {
		writeResponse(ctx, http.StatusBadRequest, "Invalid request format!", "")
		return
	}
	code, msg, userId := c.addUser(data)
	writeResponse(ctx, code, msg, userId)
}

func (c *Controller) deleteHandlerFunc(ctx *gin.Context) {
	data := ctx.PostForm("data")
	logrus.Infof("Delete User : %+v", data)

	if len(data) < 1 {
		writeResponse(ctx, http.StatusBadRequest, "Invalid request format!", "")
		return
	}
	code, msg, userId := c.deleteUser(data)
	writeResponse(ctx, code, msg, userId)
}

func (c *Controller) commitHandlerFunc(ctx *gin.Context) {
	logrus.Infof("Commit")

	code, msg := c.reloadXML()
	writeResponse(ctx, code, msg, "")
}

func (c *Controller) reloadXML() (code int, msg string) {
	cmd := fmt.Sprintf("docker exec %s sh -c \"fs_cli -x reloadxml\"", c.serverConf.ContainerName)
	data, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if err != nil {
		return 600, err.Error()
	}
	str := string(data)
	//+OK [Success]
	if strings.HasPrefix(str, "+OK") {
		return http.StatusOK, str
	}
	return 601, str
}

func (c *Controller) addUser(data string) (code int, msg string, userId string) {
	var user UserRequest

	if len(data) < 1 {
		return http.StatusBadRequest, "Invalid request format!", user.UserId
	}

	err := json.Unmarshal([]byte(data), &user)
	if err != nil || len(user.UserId) < 1 {
		return http.StatusBadRequest, "Invalid request format!", user.UserId
	}

	newXml := strings.ReplaceAll(c.serverConf.SourceXML, "$${USERID}", user.UserId)
	if len(user.Password) > 0 {
		newXml = strings.Replace(newXml, "$${default_password}", user.Password, 1)
	}
	file := fmt.Sprintf("%s.xml", user.UserId)
	file = filepath.Join(c.serverConf.FreeswitchDir, file)

	err = ioutil.WriteFile(file, []byte(newXml), 0644)
	if err != nil {
		return http.StatusInternalServerError, err.Error(), user.UserId
	}
	return http.StatusOK, "success", user.UserId
}

func (c *Controller) deleteUser(data string) (code int, msg string, userId string) {
	var user UserRequest
	err := json.Unmarshal([]byte(data), &user)
	if err != nil || len(user.UserId) < 1 {
		return http.StatusBadRequest, "Invalid request format!", user.UserId
	}

	file := fmt.Sprintf("%s.xml", user.UserId)
	file = filepath.Join(c.serverConf.FreeswitchDir, file)

	if !utils.IsFileExist(file) {
		return http.StatusNotFound, "User Not Found!", user.UserId
	}
	err = os.Remove(file)
	if err != nil {
		return http.StatusInternalServerError, err.Error(), user.UserId
	}
	return http.StatusOK, "success", user.UserId
}

func writeResponse(ctx *gin.Context, code int, msg string, userId string) {
	ctx.JSON(http.StatusOK, UserResponse{
		Code:   strconv.Itoa(code),
		Msg:    msg,
		UserId: userId,
	})
}
