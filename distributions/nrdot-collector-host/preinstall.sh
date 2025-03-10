#!/bin/sh

# Copyright The OpenTelemetry Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Create the user if NRDOT_MODE is not set to root
if [ "${NRDOT_MODE}" != "ROOT" ]; then
  getent passwd nrdot-collector-host >/dev/null || useradd --system --user-group --no-create-home --shell /sbin/nologin nrdot-collector-host
fi
