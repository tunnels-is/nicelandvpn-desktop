package core

import (
	"log"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	m "github.com/labstack/echo/v4/middleware"
)

func START_API(MONITOR chan int) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		MONITOR <- 7
	}()

	E := echo.New()

	E.Use(m.RequestLoggerWithConfig(m.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogRemoteIP: true,
		// BeforeNextFunc: func(c echo.Context) {
		// 	// log.Println(c.Request().Method, "//", c.Request().RequestURI, "(", c.Request().RemoteAddr, ")")
		// },
		LogValuesFunc: func(_ echo.Context, v m.RequestLoggerValues) error {
			log.Println(v.RoutePath)
			return nil
		},
	}))

	E.Use(m.RecoverWithConfig(m.RecoverConfig{
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   true,
		DisablePrintStack: false,
		LogLevel:          1,
	}))

	E.Use(m.SecureWithConfig(m.DefaultSecureConfig))

	corsConfig := m.CORSConfig{
		Skipper:      m.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "OPTIONS"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderCookie, echo.HeaderSetCookie, echo.HeaderXRequestedWith},
	}
	E.Use(m.CORSWithConfig(corsConfig))

	v1 := E.Group("/v1")
	// v1.Static("/", "dist")
	v1.POST("/method/:method", serveMethod)

	err := E.Start("127.0.0.1:")
	if err != nil {
		log.Println(err)
	}
}

func serveMethod(e echo.Context) error {
	method := e.Param("method")
	log.Println("METHOD:", method)
	switch method {
	case "connect":
		return HTTP_Connect(e)
	case "switch":
		return HTTP_Switch(e)
	case "disconnect":
		return HTTP_Disconnect(e)
	case "resetEverything":
		return HTTP_ResetEverything(e)
	case "setConfig":
		return HTTP_SetConfig(e)
	case "getQRCode":
		return HTTP_GetQRCode(e)
	case "getRoutersUnAuthenticated":
		return HTTP_GetRoutersUnAuthenticated(e)
	case "getRoutersAndAccessPoints":
		return HTTP_GetRoutersAndAccessPoints(e)
	case "forwardToRouter":
		return HTTP_ForwardToRouter(e)
	case "forwardToController":
		return HTTP_ForwardToController(e)
	// WS ????
	// WS ????
	// WS ????
	case "getState":
		return HTTP_GetState(e)
	case "getLoadingLogs":
	case "getLogs":
	default:
	}
	return e.JSON(200, nil)
}

func HTTP_GetState(e echo.Context) (err error) {
	PrepareState()
	return e.JSON(200, GLOBAL_STATE)
}

func HTTP_Connect(e echo.Context) (err error) {
	ns := new(CONTROLLER_SESSION_REQUEST)
	err = e.Bind(ns)
	if err != nil {
		return e.JSON(400, err)
	}

	data, code, err := Connect(ns, true)
	if err != nil {
		return e.JSON(400, err)
	}
	return e.JSON(code, data)
}

func HTTP_Switch(e echo.Context) (err error) {
	ns := new(CONTROLLER_SESSION_REQUEST)
	err = e.Bind(ns)
	if err != nil {
		return e.JSON(400, err)
	}

	data, code, err := Connect(ns, false)
	if err != nil {
		return e.JSON(400, err)
	}
	return e.JSON(code, data)
}

func HTTP_Disconnect(e echo.Context) (err error) {
	Disconnect()
	return e.JSON(200, nil)
}

func HTTP_ResetEverything(e echo.Context) (err error) {
	ResetEverything()
	return e.JSON(200, nil)
}

func HTTP_SetConfig(e echo.Context) (err error) {
	config := new(CONFIG_FORM)
	err = e.Bind(config)
	if err != nil {
		return e.JSON(400, err)
	}

	err = SetConfig(config)
	if err != nil {
		return e.JSON(400, err)
	}
	return e.JSON(200, nil)
}

func HTTP_GetQRCode(e echo.Context) (err error) {
	form := new(TWO_FACTOR_CONFIRM)
	err = e.Bind(form)
	if err != nil {
		return e.JSON(400, err)
	}
	QR, err := GetQRCode(form)
	if err != nil {
		return e.JSON(400, err)
	}
	return e.JSON(200, QR)
}

func HTTP_GetRoutersUnAuthenticated(e echo.Context) (err error) {
	data, code, err := LoadRoutersUnAuthenticated()
	if err != nil {
		return e.JSON(code, err)
	}
	return e.JSON(code, data)
}

func HTTP_GetRoutersAndAccessPoints(e echo.Context) (err error) {
	form := new(FORWARD_REQUEST)
	err = e.Bind(form)
	if err != nil {
		return e.JSON(400, err)
	}
	data, code, err := GetRoutersAndAccessPoints(form)
	if err != nil {
		return e.JSON(code, err)
	}
	return e.JSON(code, data)
}

func HTTP_ForwardToController(e echo.Context) (err error) {
	form := new(FORWARD_REQUEST)
	err = e.Bind(form)
	if err != nil {
		return e.JSON(400, err)
	}
	data, code, err := ForwardToController(form)
	if err != nil {
		return e.JSON(code, err)
	}
	return e.JSON(code, data)
}

func HTTP_ForwardToRouter(e echo.Context) (err error) {
	form := new(FORWARD_REQUEST)
	err = e.Bind(form)
	if err != nil {
		return e.JSON(400, err)
	}
	data, code, err := ForwardToRouter(form)
	if err != nil {
		return e.JSON(code, err)
	}
	return e.JSON(code, data)
}
