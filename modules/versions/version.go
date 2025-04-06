// Versions from platforms
package versions

import (
	"fmt"
	"os"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/bedrock"
	"sirherobrine23.com.br/go-bds/go-bds/java"
	"sirherobrine23.com.br/go-bds/request/v2"
)

var BedrockCacheVersion string = "https://sirherobrine23.com.br/go-bds/BedrockFetch/raw/branch/main/versions.json"

// Versions lists
var (
	BedrockVersions  bedrock.Versions
	JavaVersions     java.Versions
	PurpurVersions   java.Versions
	SpigotVersions   java.Versions
	PaperVersions    java.Versions
	FoliaVersions    java.Versions
	VelocityVersions java.Versions
)

// Time called of version fetchs
var (
	BedrockTime  time.Time
	JavaTime     time.Time
	PurpurTime   time.Time
	SpigotTime   time.Time
	PaperTime    time.Time
	FoliaTime    time.Time
	VelocityTime time.Time
)

func loopBackroud(loopTime time.Duration, fn func()) {
	fn()
	tickLoop := time.Tick(loopTime)
	for range tickLoop {
		fn()
	}
}

func init() {
	// Bedrock fetch
	go loopBackroud(time.Minute*10, func() {
		versionsBackup := BedrockVersions
		BedrockVersions = BedrockVersions[:0]
		if _, err := request.DoJSON(BedrockCacheVersion, &BedrockVersions, nil); err != nil {
			BedrockVersions = versionsBackup
			fmt.Fprintf(os.Stderr, "Failed to fetch bedrock versions: %s", err)
			return
		}
		BedrockTime = time.Now().UTC()
	})

	// Java versions
	go loopBackroud(time.Minute*40, func() {
		JavaVersions.FetchMojang()
		JavaTime = time.Now().UTC()
	})
	go loopBackroud(time.Minute*40, func() {
		SpigotVersions.FetchSpigotVersions()
		SpigotTime = time.Now().UTC()
	})
	go loopBackroud(time.Minute*40, func() {
		PurpurVersions.FetchPurpurVersions()
		PurpurTime = time.Now().UTC()
	})
	go loopBackroud(time.Minute*40, func() {
		PaperVersions.FetchPaperVersions()
		PaperTime = time.Now().UTC()
	})
	go loopBackroud(time.Minute*40, func() {
		FoliaVersions.FetchFoliaVersions()
		FoliaTime = time.Now().UTC()
	})
	go loopBackroud(time.Minute*40, func() {
		VelocityVersions.FetchVelocityVersions()
		VelocityTime = time.Now().UTC()
	})
}
