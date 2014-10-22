package jenkins

import "encoding/xml"

type Config struct {
	XMLName                          xml.Name      `xml:"project"`
	KeepDependencies                 bool          `xml:"keepDependencies"`
	Properties                       []interface{} `xml:"properties>ignored"`
	Scm                              interface{}   `xml:"scm"`
	CanRoam                          bool          `xml:"canRoam"`                          // true</canRoam>
	Disabled                         bool          `xml:"disabled"`                         // false</disabled>
	BlockBuildWhenDownstreamBuilding bool          `xml:"blockBuildWhenDownstreamBuilding"` // false</blockBuildWhenDownstreamBuilding>
	BlockBuildWhenUpstreamBuilding   bool          `xml:"blockBuildWhenUpstreamBuilding"`   // false</blockBuildWhenUpstreamBuilding>
	ConcurrentBuild                  bool          `xml:"concurrentBuild"`
	Triggers                         []interface{} `xml:"triggers>trigger"`
	Builders                         []interface{} `xml:"builders>builder"`
	AssignedToNode                   string        `xml:"assignedNode,omitempty"`
	Publishers                       []interface{} `xml:"publishers>ignored"`
	LogRotator                       *LogRotator   `xml:"logRotator,omitempty"`
}

type LogRotator struct {
	Class              string `xml:"class,attr"`
	DaysToKeep         int    `xml:"daysToKeep"`
	NumToKeep          int    `xml:"numToKeep"`
	ArtifactDaysToKeep int    `xml:"artifactDaysToKeep"`
	ArtifactNumToKeep  int    `xml:"artifactNumToKeep"`
}

func NewLogRotator(n int) *LogRotator {
	return &LogRotator{
		Class:              "hudson.tasks.LogRotator",
		DaysToKeep:         -1,
		NumToKeep:          10,
		ArtifactDaysToKeep: -1,
		ArtifactNumToKeep:  -1,
	}
}

type Builders struct {
	Tasks []interface{} `xml:""`
}

type ShellTask struct {
	XMLName xml.Name `xml:"hudson.tasks.Shell"`
	Command string   `xml:"command"`
}

type TimerTrigger struct {
	XMLName xml.Name `xml:"hudson.triggers.TimerTrigger"`
	Spec    string   `xml:"spec"`
}

func NullScm() *Scm {
	return &Scm{Class: "hudson.scm.NullSCM"}
}

type SCMTrigger struct {
	XMLName xml.Name `xml:"hudson.triggers.SCMTrigger"`
}

type Scm struct {
	Class         string `xml:"class,attr"`
	Plugin        string `xml:"plugin,attr,omitempty"`
	ConfigVersion int    `xml:"configVersion,omitempty"`
}
