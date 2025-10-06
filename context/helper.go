package context

import (
	"errors"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

// DependencyStatus struct to represent the health of a single dependency
type DependencyStatus struct {
	Status  string `json:"status"`  // "ok", "error", "degraded", "fail"
	Message string `json:"message"` // Descriptive message
}

// NewDependencyStatus creates a new instance of DependencyStatus
func NewDependencyStatus(
	status, message string) DependencyStatus {
	return DependencyStatus{
		Status:  status,
		Message: message,
	}
}

// DependencyDetails struct to represent the health of multiple dependencies
type DependencyDetails struct {
	Logger   DependencyStatus `json:"logger,omitzero"`
	Database DependencyStatus `json:"database,omitzero"`
	Nats     DependencyStatus `json:"nats,omitzero"`
	Blame    DependencyStatus `json:"blame,omitzero"`
	Paseto   DependencyStatus `json:"paseto,omitzero"`
}

func NewDependencyDetails() DependencyDetails {
	return DependencyDetails{
		Logger:   DependencyStatus{},
		Database: DependencyStatus{},
		Nats:     DependencyStatus{},
		Blame:    DependencyStatus{},
		Paseto:   DependencyStatus{},
	}
}

// CheckDependencies checks the health of the dependencies and returns the overall status
func (ctx *ServiceContext) CheckDependencies() (string, DependencyDetails) {
	var err error
	overallStatus := "OK"
	defer func() {
		if err != nil {
			overallStatus = "FAIL"
		}
	}()

	dependencyDetails := DependencyDetails{}

	logHealth := NewDependencyStatus("OK", helpers.GetHealthyMessageFor("Logger"))
	if ctx.Log == nil {
		err = errors.New("logger is empty")
		ctx.Log = log.NewBasicLogger(helpers.IsProdEnvironment(), true)
		ctx.Log.Error(constant.ControllerMessage, log.Err(err))
		logHealth.Status = "FAIL"
		logHealth.Message = err.Error()
	}
	dependencyDetails.Logger = logHealth

	if ctx.Database != nil {
		dbHealth := NewDependencyStatus("OK", helpers.GetHealthyMessageFor("Database"))
		if err1 := ctx.Database.Ping(); err1 != nil {
			err = err1
			ctx.Log.Error(constant.ControllerMessage, log.Err(err))
			dbHealth.Status = "FAIL"
			dbHealth.Message = err.Error()
		}
		dependencyDetails.Database = dbHealth
	}

	blameHealth := NewDependencyStatus("OK", helpers.GetHealthyMessageFor("Blame Wrapper"))
	if ctx.BlameWrapper == nil {
		err = errors.New("blame wrapper is empty")
		ctx.Log.Error(constant.ControllerMessage, log.Err(err))
		blameHealth.Status = "FAIL"
		blameHealth.Message = err.Error()
	}
	dependencyDetails.Blame = blameHealth

	pasetoHealth := NewDependencyStatus("OK", helpers.GetHealthyMessageFor("Paseto Wrapper"))
	if ctx.PasetoWrapper == nil {
		err = errors.New("paseto wrapper is empty")
		ctx.Log.Error(constant.ControllerMessage, log.Err(err))
		pasetoHealth.Status = "FAIL"
		pasetoHealth.Message = err.Error()
	}
	dependencyDetails.Paseto = pasetoHealth

	if ctx.NATSManager != nil {
		natsHealth := NewDependencyStatus("OK", helpers.GetHealthyMessageFor("NATS Manager"))
		if err1 := ctx.NATSManager.Ping(); err1 != nil {
			err = err1
			ctx.Log.Error(constant.ControllerMessage, log.Err(err))
			natsHealth.Status = "FAIL"
			natsHealth.Message = err.Error()
		}
		dependencyDetails.Nats = natsHealth
	}

	return overallStatus, dependencyDetails
}
