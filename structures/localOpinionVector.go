package structures

import (
	"sync"
)

type LocalOpinionVector struct {
	alpha float64
	scores [DetectionClasses]float64
	boundingBoxCoefficients [DetectionClasses]float64
	entropies [DetectionClasses]float64
	mux sync.Mutex
}

func (vector *LocalOpinionVector) getOpinionScore(class int) float64 {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	return vector.scores[class]
}

func (vector *LocalOpinionVector) getOpinionBoundingBox(class int) float64 {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	return vector.boundingBoxCoefficients[class]
}

func (vector *LocalOpinionVector) getOpinion(class int) (float64, float64, float64) {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	return vector.scores[class], vector.boundingBoxCoefficients[class], vector.entropies[class]
}

func (vector *LocalOpinionVector) updateSingleOpinion(class int, newScore, newBBcoefficient float64) {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	vector.scores[class] = vector.alpha*newScore + vector.scores[class]*(1-vector.alpha)
	vector.boundingBoxCoefficients[class] = vector.alpha*newBBcoefficient + vector.boundingBoxCoefficients[class]*(1-vector.alpha)
}

func (vector *LocalOpinionVector) updateOpinionVector(localPrediction map[int][]float64) {
	vector.mux.Lock()
	defer vector.mux.Unlock()

	for k := 1; k < DetectionClasses; k++ {
		v, exists := localPrediction[k]
		if !exists {
			v = []float64{0, 0, 1}
		}

		vector.boundingBoxCoefficients[k] = vector.alpha*v[1] + vector.boundingBoxCoefficients[k]*(1-vector.alpha)
		vector.scores[k] = vector.alpha*v[0] + vector.scores[k]*(1-vector.alpha)
		vector.entropies[k] = vector.alpha*v[2] + vector.entropies[k]*(1-vector.alpha)
	}
}