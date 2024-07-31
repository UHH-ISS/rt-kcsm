package structure

type SimplifiedUKCStage int

const (
	Recon SimplifiedUKCStage = iota
	Host
	Lateral
	Pivot
	Exfiltration
	None
)

func (stage SimplifiedUKCStage) ToUKCStages() []UKCStage {
	ukcMapping := map[SimplifiedUKCStage][]UKCStage{
		Recon:        {R, D1},
		Host:         {H},
		Lateral:      {L, S, O},
		Pivot:        {L, P, S, O},
		Exfiltration: {C, D2, E},
	}

	return ukcMapping[stage]
}

func (stage SimplifiedUKCStage) GetVictim() Victim {
	if stage == Exfiltration {
		return VictimSource
	} else {
		return VictimDestination
	}
}

func ReconStage(source IPAddress, destination IPAddress) bool {
	return !source.IsPrivate() && destination.IsPrivate()
}

func HostStage(source IPAddress, destination IPAddress) bool {
	return source.Equal(destination)
}

func PivotStage(source IPAddress, destination IPAddress) bool {
	return source.IsPrivate() && destination.IsPrivate() && !source.IsSameSubnet(destination)
}

func LateralStage(source IPAddress, destination IPAddress) bool {
	return source.IsPrivate() && destination.IsPrivate()
}

func ExfiltrationStage(source IPAddress, destination IPAddress) bool {
	return source.IsPrivate() && !destination.IsPrivate()
}

func DetermineStage(source IPAddress, destination IPAddress) SimplifiedUKCStage {
	if ReconStage(source, destination) {
		return Recon
	} else if HostStage(source, destination) {
		return Host
	} else if PivotStage(source, destination) {
		return Pivot
	} else if LateralStage(source, destination) {
		return Lateral
	} else if ExfiltrationStage(source, destination) {
		return Exfiltration
	} else {
		return None
	}
}
