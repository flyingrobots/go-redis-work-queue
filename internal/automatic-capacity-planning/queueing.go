// Copyright 2025 James Ross
package capacityplanning

import (
	"math"
	"time"
)

// QueueingCalculator interface defines queueing theory calculations
type QueueingCalculator interface {
	Calculate(lambda, mu float64, servers int, metrics Metrics) *QueueingResult
	CalculateCapacity(lambda, mu float64, targetLatency time.Duration) int
	EstimateServiceRate(metrics Metrics) float64
}

// queueingCalculator implements queueing theory models
type queueingCalculator struct {
	config PlannerConfig
}

// NewQueueingCalculator creates a new queueing calculator
func NewQueueingCalculator(config PlannerConfig) QueueingCalculator {
	return &queueingCalculator{
		config: config,
	}
}

// Calculate performs queueing theory calculations based on the configured model
func (q *queueingCalculator) Calculate(lambda, mu float64, servers int, metrics Metrics) *QueueingResult {
	switch q.config.QueueingModel {
	case "mm1":
		return q.calculateMM1(lambda, mu, metrics)
	case "mmc":
		return q.calculateMMC(lambda, mu, servers, metrics)
	case "mgc":
		return q.calculateMGC(lambda, mu, servers, metrics)
	default:
		// Default to M/M/c model
		return q.calculateMMC(lambda, mu, servers, metrics)
	}
}

// calculateMM1 implements the M/M/1 queueing model (single server)
func (q *queueingCalculator) calculateMM1(lambda, mu float64, metrics Metrics) *QueueingResult {
	// M/M/1 model assumes single server
	servers := 1

	// Utilization ρ = λ/μ
	rho := lambda / mu

	// System is unstable if ρ >= 1
	if rho >= 1.0 {
		return &QueueingResult{
			Utilization:  1.0,
			QueueLength:  math.Inf(1),
			WaitTime:     time.Duration(math.Inf(1)),
			ResponseTime: time.Duration(math.Inf(1)),
			Throughput:   mu,
			Capacity:     servers,
			Model:        "M/M/1",
			Confidence:   0.0,
			Assumptions:  []string{"System unstable: λ >= μ"},
		}
	}

	// Queue length L_q = ρ²/(1-ρ)
	queueLength := (rho * rho) / (1.0 - rho)

	// Wait time W_q = L_q/λ
	waitTimeSeconds := queueLength / lambda
	waitTime := time.Duration(waitTimeSeconds * float64(time.Second))

	// Response time W = W_q + 1/μ
	responseTimeSeconds := waitTimeSeconds + (1.0 / mu)
	responseTime := time.Duration(responseTimeSeconds * float64(time.Second))

	// Throughput (effective arrival rate)
	throughput := lambda // In stable system, throughput = arrival rate

	// Calculate confidence based on model assumptions
	confidence := q.calculateModelConfidence("M/M/1", lambda, mu, servers, metrics)

	return &QueueingResult{
		Utilization:  rho,
		QueueLength:  queueLength,
		WaitTime:     waitTime,
		ResponseTime: responseTime,
		Throughput:   throughput,
		Capacity:     servers,
		Model:        "M/M/1",
		Confidence:   confidence,
		Assumptions:  []string{"Poisson arrivals", "Exponential service times", "Single server", "FIFO discipline"},
	}
}

// calculateMMC implements the M/M/c queueing model (multiple servers)
func (q *queueingCalculator) calculateMMC(lambda, mu float64, servers int, metrics Metrics) *QueueingResult {
	c := float64(servers)

	// Total service rate
	totalMu := c * mu

	// Utilization ρ = λ/(c×μ)
	rho := lambda / totalMu

	// System is unstable if ρ >= 1
	if rho >= 1.0 {
		return &QueueingResult{
			Utilization:  1.0,
			QueueLength:  math.Inf(1),
			WaitTime:     time.Duration(math.Inf(1)),
			ResponseTime: time.Duration(math.Inf(1)),
			Throughput:   totalMu,
			Capacity:     servers,
			Model:        "M/M/c",
			Confidence:   0.0,
			Assumptions:  []string{"System unstable: λ >= c×μ"},
		}
	}

	// Calculate P_0 (probability of empty system)
	// P_0 = [Σ(n=0 to c-1)((λ/μ)^n/n!) + (λ/μ)^c/(c!(1-ρ))]^(-1)
	p0 := q.calculateP0MMC(lambda, mu, servers)

	// Traffic intensity a = λ/μ
	a := lambda / mu

	// Queue length (jobs waiting in queue, not being served)
	// L_q = (P_0 × (a^c) × ρ) / (c! × (1-ρ)²)
	factorial_c := factorial(servers)
	lq := (p0 * math.Pow(a, c) * rho) / (factorial_c * math.Pow(1.0-rho, 2))

	// Wait time W_q = L_q/λ
	waitTimeSeconds := lq / lambda
	waitTime := time.Duration(waitTimeSeconds * float64(time.Second))

	// Response time W = W_q + 1/μ
	responseTimeSeconds := waitTimeSeconds + (1.0 / mu)
	responseTime := time.Duration(responseTimeSeconds * float64(time.Second))

	// Throughput
	throughput := lambda

	// Calculate confidence
	confidence := q.calculateModelConfidence("M/M/c", lambda, mu, servers, metrics)

	return &QueueingResult{
		Utilization:  rho,
		QueueLength:  lq,
		WaitTime:     waitTime,
		ResponseTime: responseTime,
		Throughput:   throughput,
		Capacity:     servers,
		Model:        "M/M/c",
		Confidence:   confidence,
		Assumptions:  []string{"Poisson arrivals", "Exponential service times", "Multiple servers", "FIFO discipline"},
	}
}

// calculateMGC implements the M/G/c queueing model (general service time distribution)
func (q *queueingCalculator) calculateMGC(lambda, mu float64, servers int, metrics Metrics) *QueueingResult {
	// First calculate M/M/c result as baseline
	mmcResult := q.calculateMMC(lambda, mu, servers, metrics)

	// Apply Pollaczek-Khinchin correction for general service time
	// W_q(M/G/c) ≈ (C_s² + 1)/2 × W_q(M/M/c)
	// where C_s² is the squared coefficient of variation of service time

	serviceTimeMean := 1.0 / mu
	serviceTimeStd := metrics.ServiceTimeStd.Seconds()

	// Coefficient of variation squared
	cs2 := math.Pow(serviceTimeStd/serviceTimeMean, 2)

	// Correction factor
	correctionFactor := (cs2 + 1.0) / 2.0

	// Apply correction to wait time
	correctedWaitTime := time.Duration(float64(mmcResult.WaitTime) * correctionFactor)
	correctedResponseTime := correctedWaitTime + time.Duration(serviceTimeMean*float64(time.Second))

	// Recalculate queue length using Little's Law
	correctedQueueLength := lambda * correctedWaitTime.Seconds()

	// Adjust confidence based on how well the general distribution assumption fits
	confidence := mmcResult.Confidence * q.calculateGeneralServiceConfidence(metrics)

	assumptions := append(mmcResult.Assumptions, "General service time distribution")
	// Replace exponential service time assumption
	for i, assumption := range assumptions {
		if assumption == "Exponential service times" {
			assumptions[i] = "General service time distribution"
			break
		}
	}

	return &QueueingResult{
		Utilization:  mmcResult.Utilization,
		QueueLength:  correctedQueueLength,
		WaitTime:     correctedWaitTime,
		ResponseTime: correctedResponseTime,
		Throughput:   mmcResult.Throughput,
		Capacity:     servers,
		Model:        "M/G/c",
		Confidence:   confidence,
		Assumptions:  assumptions,
	}
}

// calculateP0MMC calculates the probability of empty system for M/M/c
func (q *queueingCalculator) calculateP0MMC(lambda, mu float64, servers int) float64 {
	a := lambda / mu // Traffic intensity
	c := float64(servers)

	// First sum: Σ(n=0 to c-1)((a^n)/n!)
	sum1 := 0.0
	for n := 0; n < servers; n++ {
		sum1 += math.Pow(a, float64(n)) / factorial(n)
	}

	// Second term: (a^c)/(c!(1-ρ))
	rho := lambda / (c * mu)
	if rho >= 1.0 {
		return 0.0 // System unstable
	}

	factorial_c := factorial(servers)
	sum2 := math.Pow(a, c) / (factorial_c * (1.0 - rho))

	// P_0 = 1 / (sum1 + sum2)
	return 1.0 / (sum1 + sum2)
}

// CalculateCapacity determines the minimum number of servers needed for target latency
func (q *queueingCalculator) CalculateCapacity(lambda, mu float64, targetLatency time.Duration) int {
	targetSeconds := targetLatency.Seconds()

	// Start with minimum and increase until we meet the target
	for servers := 1; servers <= 1000; servers++ { // Reasonable upper bound
		result := q.Calculate(lambda, mu, servers, Metrics{})

		if result.ResponseTime.Seconds() <= targetSeconds {
			return servers
		}

		// If utilization is still >= 1, system is unstable
		if result.Utilization >= 1.0 {
			continue
		}
	}

	// If we can't meet the target with 1000 servers, return 1000
	return 1000
}

// EstimateServiceRate estimates the service rate from metrics
func (q *queueingCalculator) EstimateServiceRate(metrics Metrics) float64 {
	if metrics.ServiceTime <= 0 {
		return 0.0
	}

	// Service rate μ = 1 / service_time
	return 1.0 / metrics.ServiceTime.Seconds()
}

// calculateModelConfidence estimates how well the model fits the actual system
func (q *queueingCalculator) calculateModelConfidence(model string, lambda, mu float64, servers int, metrics Metrics) float64 {
	baseConfidence := 0.8 // Start with 80% confidence

	// Reduce confidence if utilization is very high
	rho := lambda / (float64(servers) * mu)
	if rho > 0.9 {
		baseConfidence *= 0.8
	} else if rho > 0.8 {
		baseConfidence *= 0.9
	}

	// Reduce confidence for M/M/1 when we actually have multiple workers
	if model == "M/M/1" && servers > 1 {
		baseConfidence *= 0.7
	}

	// Increase confidence for M/G/c when service time variance is high
	if model == "M/G/c" {
		serviceTimeMean := 1.0 / mu
		serviceTimeStd := metrics.ServiceTimeStd.Seconds()
		cv := serviceTimeStd / serviceTimeMean

		if cv > 0.5 { // High variability favors M/G/c
			baseConfidence *= 1.1
		}
	}

	// Ensure confidence is between 0 and 1
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	if baseConfidence < 0.1 {
		baseConfidence = 0.1
	}

	return baseConfidence
}

// calculateGeneralServiceConfidence adjusts confidence for general service time distribution
func (q *queueingCalculator) calculateGeneralServiceConfidence(metrics Metrics) float64 {
	if metrics.ServiceTimeStd <= 0 {
		return 0.7 // Low confidence without variance data
	}

	serviceTimeMean := metrics.ServiceTime.Seconds()
	serviceTimeStd := metrics.ServiceTimeStd.Seconds()

	// Coefficient of variation
	cv := serviceTimeStd / serviceTimeMean

	// Higher CV favors general distribution model
	if cv > 1.0 {
		return 0.95
	} else if cv > 0.5 {
		return 0.85
	} else if cv < 0.2 {
		return 0.75 // Low variance might be closer to exponential
	}

	return 0.8
}

// Utility functions

// factorial calculates n! (with memoization for efficiency)
var factorialCache = map[int]float64{
	0: 1,
	1: 1,
}

func factorial(n int) float64 {
	if n < 0 {
		return 0
	}

	if val, exists := factorialCache[n]; exists {
		return val
	}

	result := float64(n) * factorial(n-1)
	factorialCache[n] = result
	return result
}

// erlangC calculates the Erlang C formula for M/M/c systems
func erlangC(lambda, mu float64, servers int) float64 {
	a := lambda / mu // Traffic intensity
	c := float64(servers)

	if lambda >= c*mu {
		return 1.0 // System overloaded
	}

	// Calculate using the standard Erlang C formula
	numerator := math.Pow(a, c) / factorial(servers)
	denominator := 0.0

	// Sum for k=0 to c-1
	for k := 0; k < servers; k++ {
		denominator += math.Pow(a, float64(k)) / factorial(k)
	}

	// Add the c-th term
	rho := lambda / (c * mu)
	denominator += (math.Pow(a, c) / factorial(servers)) / (1.0 - rho)

	return numerator / denominator
}

// littlesLaw applies Little's Law: L = λW
func littlesLaw(lambda float64, waitTime time.Duration) float64 {
	return lambda * waitTime.Seconds()
}

// jacksonNetwork handles multiple queue networks (future enhancement)
func (q *queueingCalculator) analyzeNetwork(queues []Metrics) *QueueingResult {
	// Placeholder for Jackson network analysis
	// This would handle multiple interconnected queues
	return nil
}