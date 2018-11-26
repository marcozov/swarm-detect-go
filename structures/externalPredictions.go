package structures

func (node *Node) InitExternalPredictionsMap() *SafeMapSinglePredictions {
	peers := &SafeMapSinglePredictions{
		v: make(map[string]SinglePrediction),
	}

	return peers
}

func (node *Node) GetPrediction(host string) *SinglePrediction {
	return node.ExternalPredictions.getPrediction(host)
}


func (node *Node) RemovePrediction(host string) {
	node.ExternalPredictions.removePrediction(host)
}
