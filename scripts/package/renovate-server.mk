# Copyright 2020 The arhat.dev Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# linux
#
package.renovate-server.deb.amd64:
	sh scripts/package/package.sh $@

package.renovate-server.deb.armv6:
	sh scripts/package/package.sh $@

package.renovate-server.deb.armv7:
	sh scripts/package/package.sh $@

package.renovate-server.deb.arm64:
	sh scripts/package/package.sh $@

package.renovate-server.deb.all: \
	package.renovate-server.deb.amd64 \
	package.renovate-server.deb.armv6 \
	package.renovate-server.deb.armv7 \
	package.renovate-server.deb.arm64

package.renovate-server.rpm.amd64:
	sh scripts/package/package.sh $@

package.renovate-server.rpm.armv7:
	sh scripts/package/package.sh $@

package.renovate-server.rpm.arm64:
	sh scripts/package/package.sh $@

package.renovate-server.rpm.all: \
	package.renovate-server.rpm.amd64 \
	package.renovate-server.rpm.armv7 \
	package.renovate-server.rpm.arm64

package.renovate-server.linux.all: \
	package.renovate-server.deb.all \
	package.renovate-server.rpm.all

#
# windows
#

package.renovate-server.msi.amd64:
	sh scripts/package/package.sh $@

package.renovate-server.msi.arm64:
	sh scripts/package/package.sh $@

package.renovate-server.msi.all: \
	package.renovate-server.msi.amd64 \
	package.renovate-server.msi.arm64

package.renovate-server.windows.all: \
	package.renovate-server.msi.all

#
# darwin
#

package.renovate-server.pkg.amd64:
	sh scripts/package/package.sh $@

package.renovate-server.pkg.arm64:
	sh scripts/package/package.sh $@

package.renovate-server.pkg.all: \
	package.renovate-server.pkg.amd64 \
	package.renovate-server.pkg.arm64

package.renovate-server.darwin.all: \
	package.renovate-server.pkg.all
