// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"gopkg.in/yaml.v3"

	"github.com/newrelic/opentelemetry-collector-releases/cmd/goreleaser/internal"
)

var distFlag = flag.String("d", "", "Collector distributions to build")
var nightlyFlag = flag.Bool("n", false, "Whether we're building a nightly config")

func main() {
	flag.Parse()

	if len(*distFlag) == 0 {
		log.Fatal("no distribution to build")
	}

	var project config.Project

	project = internal.Generate(*distFlag, *nightlyFlag)

	e := yaml.NewEncoder(os.Stdout)
	e.SetIndent(2)
	if err := e.Encode(&project); err != nil {
		log.Fatal(err)
	}
}
