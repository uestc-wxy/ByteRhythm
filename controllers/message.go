package controllers

import (
	"ByteRhythm/models"
	"ByteRhythm/object"
	"ByteRhythm/utils"
	"github.com/beego/beego/v2/client/orm"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// MessageController operations for Message
type MessageController struct {
	baseController
}

func (c *MessageController) ChatMessage() {
	token := c.GetString("token")
	toUserId, _ := c.GetInt("to_user_id")

	fromUserId, _ := utils.GetUserIdFromToken(token)

	if fromUserId == toUserId {
		c.Data["json"] = map[string]interface{}{
			"status_code": 1,
			"status_msg":  "不能查看与自己的聊天记录",
		}
		c.ServeJSON()
		return
	}
	messageList, err := GetALLMessage(c, fromUserId, toUserId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"status_code": 1,
			"status_msg":  err.Error(),
		}
		c.ServeJSON()
		return
	}
	c.Data["json"] = map[string]interface{}{
		"status_code":  "0",
		"status_msg":   "获取聊天记录成功！",
		"message_list": messageList,
	}
	c.ServeJSON()

	return

}

func GetALLMessage(c *MessageController, fromUseId int, toUseId int) (messageList []object.MessageDto, err error) {
	var maps []orm.Params

	//  查询互发的聊天记录
	_, err = c.o.Raw(`select * from message where (from_user_id = ? and to_user_id = ?) or (from_user_id = ? and to_user_id = ?) `).SetArgs(fromUseId, toUseId, toUseId, fromUseId).Values(&maps)
	if err != nil {
		return nil, err
	}

	for i := range maps {

		message := maps[i]

		id, _ := strconv.Atoi(message["id"].(string))
		from, _ := strconv.Atoi(message["from_user_id"].(string))
		to, _ := strconv.Atoi(message["to_user_id"].(string))

		parseTime, _ := time.ParseInLocation("2006-01-02 15:04:05", message["create_time"].(string), time.Local)

		// 转化为我们需要的格式
		ms := object.MessageDto{Id: id, FromUserId: from, ToUserId: to, Content: message["content"].(string), CreateTime: int(parseTime.Unix())}

		messageList = append(messageList, ms)
	}
	return messageList, nil
}

func (c *MessageController) ActionMessage() {
	token := c.GetString("token")
	from_user_id, err := utils.GetUserIdFromToken(token)
	if err != nil {
		return
	}

	actionType := c.GetString("action_type")
	if actionType == "1" {
		toUserId, err := strconv.Atoi(c.GetString("to_user_id"))
		if err != nil {
			c.handleError(err)
			return
		}

		content := c.GetString("content")

		user := &models.User{Id: from_user_id}
		if err = c.o.Read(user); err != nil {
			c.handleError(err)
			return
		}
		toUser := &models.User{Id: toUserId}
		if err = c.o.Read(toUser); err != nil {
			c.handleError(err)
			return
		}

		var message = models.Message{
			FromUserId: user,
			ToUserId:   toUser,
			Content:    content,
		}
		_, err = c.o.Insert(&message)
		if err != nil {
			c.handleError(err)
			return
		}
		c.Data["json"] = map[string]interface{}{
			"status_code": 0,
			"status_msg":  "发送消息成功！",
		}
		c.ServeJSON()
	} else {
		err := errors.New("非发送消息操作！")
		c.handleError(err)
		return
	}
}

func (c *MessageController) handleError(err error) {
	c.Data["json"] = map[string]interface{}{
		"status_code": 1,
		"status_msg":  err.Error(),
	}
	c.ServeJSON()
}
