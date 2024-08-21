// Versions from platforms
package versions

import (
	"sync"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/bedrock"
	"sirherobrine23.com.br/go-bds/go-bds/java"
	"sirherobrine23.com.br/go-bds/go-bds/semver"
)

// Version fetch next time
var (
	NextFetchBedrock time.Time
	NextFetchJava    time.Time

	onceInit *sync.Once = &sync.Once{}
)

// Versions lists
var (
	BedrockVersions bedrock.Versions
	JavaVersions    java.Versions
	PurpurVersions  java.Versions
)

// Start fetch versions
func InitFetch() {
	tick := time.Minute * 5
	onceInit.Do(func() {
		go func() {
			err := error(nil)
			if BedrockVersions, err = bedrock.FromVersions(); err != nil {
				println(err.Error())
			}
			NextFetchBedrock = time.Now().UTC().Add(tick)

			for range time.Tick(tick) {
				if BedrockVersions, err = bedrock.FromVersions(); err != nil {
					println(err.Error())
				}
				NextFetchBedrock = time.Now().UTC()
			}
		}()
		go func() {
			err := error(nil)
			if JavaVersions, err = java.ListMojang(); err != nil {
				println(err.Error())
			}
			semver.SortStruct(JavaVersions)

			if PurpurVersions, err = java.ListPurpur(); err != nil {
				println(err.Error())
			}
			semver.SortStruct(PurpurVersions)
			NextFetchJava = time.Now().UTC().Add(tick)

			for range time.Tick(tick) {
				err := error(nil)
				if JavaVersions, err = java.ListMojang(); err != nil {
					println(err.Error())
				}
				semver.SortStruct(JavaVersions)

				if PurpurVersions, err = java.ListPurpur(); err != nil {
					println(err.Error())
				}
				semver.SortStruct(PurpurVersions)
				NextFetchJava = time.Now().UTC()
			}
		}()
	})
}
