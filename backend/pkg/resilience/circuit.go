package resilience

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/go-micro/plugins/v4/wrapper/breaker/hystrix"
)

func BuildHystrixCommandConfig(resilienceConfig *config.ResilienceConfig) hystrix.CommandConfig {
	var config hystrix.CommandConfig
	if resilienceConfig.Resilience.CircuitBreaker.Timeout > 0 {
		config.Timeout = resilienceConfig.Resilience.CircuitBreaker.Timeout
	}

	if resilienceConfig.Resilience.CircuitBreaker.MaxConcurrent > 0 {
		config.MaxConcurrentRequests = resilienceConfig.Resilience.CircuitBreaker.MaxConcurrent
	}

	if resilienceConfig.Resilience.CircuitBreaker.VolumeThreshold > 0 {
		config.RequestVolumeThreshold = resilienceConfig.Resilience.CircuitBreaker.VolumeThreshold
	}

	if resilienceConfig.Resilience.CircuitBreaker.SleepWindow > 0 {
		config.SleepWindow = resilienceConfig.Resilience.CircuitBreaker.SleepWindow
	}

	if resilienceConfig.Resilience.CircuitBreaker.ErrorPercentThreshold > 0 {
		config.ErrorPercentThreshold = resilienceConfig.Resilience.CircuitBreaker.ErrorPercentThreshold
	}

	return config
}
