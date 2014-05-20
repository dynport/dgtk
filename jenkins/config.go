package jenkins

import (
	"encoding/xml"
)

type Config struct {
	XMLName                          xml.Name      `xml:"project"`
	KeepDependencies                 bool          `xml:"keepDependencies"`
	Properties                       []interface{} `xml:"properties"`
	CanRoam                          bool          `xml:"canRoam"`                          // true</canRoam>
	Disabled                         bool          `xml:"disabled"`                         // false</disabled>
	BlockBuildWhenDownstreamBuilding bool          `xml:"blockBuildWhenDownstreamBuilding"` // false</blockBuildWhenDownstreamBuilding>
	BlockBuildWhenUpstreamBuilding   bool          `xml:"blockBuildWhenUpstreamBuilding"`   // false</blockBuildWhenUpstreamBuilding>
	ConcurrentBuild                  bool          `xml:"concurrentBuild"`
}

type SCMTrigger struct {
	XMLName xml.Name `xml:"hudson.triggers.SCMTrigger"`
}

type Scm struct {
	Class         string `xml:"class,attr"`
	Plugin        string `xml:"plugin,attr"`
	ConfigVersion int    `xml:"configVersion"`
}
