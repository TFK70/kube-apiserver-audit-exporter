package collector

import (
	"github.com/labstack/echo/v4"

	vm "github.com/VictoriaMetrics/metrics"
)

func HandleMetrics(c echo.Context) error {
	vm.WritePrometheus(c.Response().Writer, true)
	return nil
}
