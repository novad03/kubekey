package prepare

// Condition struct is a Default implementation.
type Condition struct {
	BasePrepare
	Cond bool
}

func (c *Condition) PreCheck() (bool, error) {
	if c.Cond {
		return true, nil
	}
	return false, nil
}

type OnlyFirstMaster struct {
	BasePrepare
}

func (o *OnlyFirstMaster) PreCheck() (bool, error) {
	if o.Runtime.Runner.Host.IsMaster && o.Runtime.Runner.Host.Name == o.Runtime.MasterNodes[0].Name {
		return true, nil
	}
	return false, nil
}

type OnlyWorker struct {
	BasePrepare
}

func (o *OnlyWorker) PreCheck() (bool, error) {
	if o.Runtime.Runner.Host.IsWorker && !o.Runtime.Runner.Host.IsMaster {
		return true, nil
	}
	return false, nil
}

type OnlyK3s struct {
	BasePrepare
}

func (o *OnlyK3s) PreCheck() (bool, error) {
	if o.Runtime.Cluster.Kubernetes.Type == "k3s" {
		return true, nil
	}
	return false, nil
}

type OnlyKubernetes struct {
	BasePrepare
}

func (o *OnlyKubernetes) PreCheck() (bool, error) {
	if o.Runtime.Cluster.Kubernetes.Type != "k3s" {
		return true, nil
	}
	return false, nil
}
