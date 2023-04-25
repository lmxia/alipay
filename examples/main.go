package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/xid"
	"log"
	"net/http"
)

var aliClient *alipay.Client

const (
	kAppId      = "2021000122663656"
	kPrivateKey = "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCnHOjy1OgRtt9TJu+tYt4zOrDdf45akyJk35PxPRBAh03GelWakKRO7/URD0Z1x1YNCPgxBthXUzZsPjmewuuLtznX9aVVPu4MNwP1EfozJ+n0Kn3fawzhdO5MKBd5eumALQ1+HQ1fq2Kl9tVfgHLRZg/JRB8cPlGgx+fSHFfnBiBaM1Sa+xP9y+3+GsGloRpIA95WBGCJIPaTkOx52wzvXSwg6PB0kzq3nLlIEEQ0Wz5XRUwlw8Y7FIg2IWEkAyEYWTHkL5rzZxreASxMOoS7uVnDvSAaPaPUfpODwhJViOFzMjOqjf19kfbg5EhGI0O1mHmBjh4/97NO3W2gKSH7AgMBAAECggEBAIZwQItScpwFpWXstfajyiQmfDDFNE1zfsDuCMOTH2b6MryQoBtnb3e9nUarJkEMKxLze09dtV+TJv4vlQU+xGWy8orlKKwKo1EaVUmru7+5xYHTSU4afdNG0Ypc2n21PvIJzIf/cUncw9DGYWOiHzyMQfjln62GCP8yszGZ8bF9NDV58Vv53+LVoAeTv/JA2EXKiHLrZvf8REm23bTT7SCHTq0N9M6rIbuiZ2OiBOL10HQVuG0RCd/pY/9hlMO0GX8HkT77EN0xO62cba+mfozerHTOU6XgSszWEjdzlX+daggiUrOr3kPaKxSfZVtr50HY8Er7V6M55H0fpMBbUtkCgYEA1q5wcDzTAMHd0Lejs5jK2XVF+DoNTn1VtxVA/w3J376UBU9hjcJNutCIL4uVB7T2iZbSGV5NTR67VEiaMx/X4WiRkftG4H0BH7oJ/JVA0tNi6cLY1lGjUU9sP4IjdqN5OcKudOKFbrtxDO8jxhb8h2k8t3kBknxtn7HnTXCpsycCgYEAx0a515LjCNW7iNqsLVWkbX3HsP30Hm0Ghz6Dod77D0Szf+qV2dbWy64EVvnakCi80Cqxm/xY7CJtI+aWLYn/q0pLFTjzmwf6O1Yw5y6SaoTu9ECQV8402PIOLmW+B18iJFMJ1R30asUs/qx8KqJlm2Iz4iOHPtqP13AiWw3uTw0CgYALWAZa9+vSY2wcJkgBKna1jOvYlQC1AAxycy4PDCR5rTFXIn2uJvFCiNhZYs/KK3bHiG+rpX8CLziI2JlFUE5w+7yNcTCqlhBkI8l5Tk2xljfo0EHy+TdfCYpXxMGo+DRkp/Pd+0Y/tqnfnBdQ1VAcu6PYsg0yN173jEgDoItnCwKBgH8DAjJ3icM4zxXUIoemnW24DI4v3ueBn/aVjrqeb7B0nl/6edZli3Q4jsWM7JSTknyvqZJ9TYP8EUofjzqpSU64xJBbQ1FdzU0Ci5rd4S6JWfBOMnH0mVRpO0axTGRQa2dxkcPHGuDumdYcw+s8pLxb5CvPb0VNcv7iltMoVusFAoGAHyd01QeWiT8xUH/qKYrv+hRkXgAgITB+MnoYZ87h3DSFAyQ1o391qqIaNHoCtQViDaRMop239EpRKGlHNZIHfCpqhVN0qjRhI3X1Vpnmm4pjihORYimJcKQvVwEojikoXWeKeycOica6Kor3YShSPbMJYQqnBo6c7v6NR/FipEE="
	kServerPort = "9989"
	// TODO 设置回调地址域名
	kServerDomain = "127.0.0.1"
)

func main() {
	var err error

	if aliClient, err = alipay.New(kAppId, kPrivateKey, false); err != nil {
		log.Println("初始化支付宝失败", err)
		return
	}

	// 加载证书
	if err = aliClient.LoadAppPublicCertFromFile("appPublicCert.cer"); err != nil {
		log.Println("加载证书发生错误", err)
		return
	}
	if err = aliClient.LoadAliPayRootCertFromFile("alipayRootCert.cer"); err != nil {
		log.Println("加载证书发生错误", err)
		return
	}
	if err = aliClient.LoadAliPayPublicCertFromFile("alipayPublicCert.cer"); err != nil {
		log.Println("加载证书发生错误", err)
		return
	}

	if err = aliClient.SetEncryptKey("XaZdAx6AyH7M51AZRKjJoQ=="); err != nil {
		log.Println("加载内容加密密钥发生错误", err)
		return
	}

	var s = gin.Default()
	s.GET("/alipay/pay", pay)
	s.GET("/alipay/callback", callback)
	s.POST("/alipay/notify", notify)
	s.Run(":" + kServerPort)
}

func pay(c *gin.Context) {
	var tradeNo = fmt.Sprintf("%d", xid.Next())

	var p = alipay.TradePagePay{}
	p.NotifyURL = kServerDomain + "/alipay/notify"
	p.ReturnURL = kServerDomain + "/alipay/callback"
	p.Subject = "支付测试:" + tradeNo
	p.OutTradeNo = tradeNo
	p.TotalAmount = "10.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, _ := aliClient.TradePagePay(p)

	c.Redirect(http.StatusTemporaryRedirect, url.String())
}

func callback(c *gin.Context) {
	c.Request.ParseForm()

	ok, err := aliClient.VerifySign(c.Request.Form)
	if err != nil {
		log.Println("回调验证签名发生错误", err)
		c.String(http.StatusBadRequest, "回调验证签名发生错误")
		return
	}

	if ok == false {
		log.Println("回调验证签名未通过")
		c.String(http.StatusBadRequest, "回调验证签名未通过")
		return
	}

	log.Println("回调验证签名通过")

	var outTradeNo = c.Request.Form.Get("out_trade_no")
	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo
	rsp, err := aliClient.TradeQuery(p)
	if err != nil {
		c.String(http.StatusBadRequest, "验证订单 %s 信息发生错误: %s", outTradeNo, err.Error())
		return
	}
	if rsp.IsSuccess() == false {
		c.String(http.StatusBadRequest, "验证订单 %s 信息发生错误: %s-%s", outTradeNo, rsp.Content.Msg, rsp.Content.SubMsg)
		return
	}

	c.String(http.StatusOK, "订单 %s 支付成功", outTradeNo)
}

func notify(c *gin.Context) {
	c.Request.ParseForm()

	ok, err := aliClient.VerifySign(c.Request.Form)
	if err != nil {
		log.Println("异步通知验证签名发生错误", err)
		return
	}

	if ok == false {
		log.Println("异步通知验证签名未通过")
		return
	}

	log.Println("异步通知验证签名通过")

	var outTradeNo = c.Request.Form.Get("out_trade_no")
	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo
	rsp, err := aliClient.TradeQuery(p)
	if err != nil {
		log.Printf("异步通知验证订单 %s 信息发生错误: %s \n", outTradeNo, err.Error())
		return
	}
	if rsp.IsSuccess() == false {
		log.Printf("异步通知验证订单 %s 信息发生错误: %s-%s \n", outTradeNo, rsp.Content.Msg, rsp.Content.SubMsg)
		return
	}

	log.Printf("订单 %s 支付成功 \n", outTradeNo)

	aliClient.ACKNotification(c.Writer)
}
