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

# native
renovate-server:
	sh scripts/build/build.sh $@

# linux
renovate-server.linux.x86:
	sh scripts/build/build.sh $@

renovate-server.linux.amd64:
	sh scripts/build/build.sh $@

renovate-server.linux.armv5:
	sh scripts/build/build.sh $@

renovate-server.linux.armv6:
	sh scripts/build/build.sh $@

renovate-server.linux.armv7:
	sh scripts/build/build.sh $@

renovate-server.linux.arm64:
	sh scripts/build/build.sh $@

renovate-server.linux.mips:
	sh scripts/build/build.sh $@

renovate-server.linux.mipshf:
	sh scripts/build/build.sh $@

renovate-server.linux.mipsle:
	sh scripts/build/build.sh $@

renovate-server.linux.mipslehf:
	sh scripts/build/build.sh $@

renovate-server.linux.mips64:
	sh scripts/build/build.sh $@

renovate-server.linux.mips64hf:
	sh scripts/build/build.sh $@

renovate-server.linux.mips64le:
	sh scripts/build/build.sh $@

renovate-server.linux.mips64lehf:
	sh scripts/build/build.sh $@

renovate-server.linux.ppc64:
	sh scripts/build/build.sh $@

renovate-server.linux.ppc64le:
	sh scripts/build/build.sh $@

renovate-server.linux.s390x:
	sh scripts/build/build.sh $@

renovate-server.linux.riscv64:
	sh scripts/build/build.sh $@

renovate-server.linux.all: \
	renovate-server.linux.x86 \
	renovate-server.linux.amd64 \
	renovate-server.linux.armv5 \
	renovate-server.linux.armv6 \
	renovate-server.linux.armv7 \
	renovate-server.linux.arm64 \
	renovate-server.linux.mips \
	renovate-server.linux.mipshf \
	renovate-server.linux.mipsle \
	renovate-server.linux.mipslehf \
	renovate-server.linux.mips64 \
	renovate-server.linux.mips64hf \
	renovate-server.linux.mips64le \
	renovate-server.linux.mips64lehf \
	renovate-server.linux.ppc64 \
	renovate-server.linux.ppc64le \
	renovate-server.linux.s390x \
	renovate-server.linux.riscv64

renovate-server.darwin.amd64:
	sh scripts/build/build.sh $@

# # currently darwin/arm64 build will fail due to golang link error
# renovate-server.darwin.arm64:
# 	sh scripts/build/build.sh $@

renovate-server.darwin.all: \
	renovate-server.darwin.amd64

renovate-server.windows.x86:
	sh scripts/build/build.sh $@

renovate-server.windows.amd64:
	sh scripts/build/build.sh $@

renovate-server.windows.armv5:
	sh scripts/build/build.sh $@

renovate-server.windows.armv6:
	sh scripts/build/build.sh $@

renovate-server.windows.armv7:
	sh scripts/build/build.sh $@

# # currently no support for windows/arm64
# renovate-server.windows.arm64:
# 	sh scripts/build/build.sh $@

renovate-server.windows.all: \
	renovate-server.windows.x86 \
	renovate-server.windows.amd64

# # android build requires android sdk
# renovate-server.android.amd64:
# 	sh scripts/build/build.sh $@

# renovate-server.android.x86:
# 	sh scripts/build/build.sh $@

# renovate-server.android.armv5:
# 	sh scripts/build/build.sh $@

# renovate-server.android.armv6:
# 	sh scripts/build/build.sh $@

# renovate-server.android.armv7:
# 	sh scripts/build/build.sh $@

# renovate-server.android.arm64:
# 	sh scripts/build/build.sh $@

# renovate-server.android.all: \
# 	renovate-server.android.amd64 \
# 	renovate-server.android.arm64 \
# 	renovate-server.android.x86 \
# 	renovate-server.android.armv7 \
# 	renovate-server.android.armv5 \
# 	renovate-server.android.armv6

renovate-server.freebsd.amd64:
	sh scripts/build/build.sh $@

renovate-server.freebsd.x86:
	sh scripts/build/build.sh $@

renovate-server.freebsd.armv5:
	sh scripts/build/build.sh $@

renovate-server.freebsd.armv6:
	sh scripts/build/build.sh $@

renovate-server.freebsd.armv7:
	sh scripts/build/build.sh $@

renovate-server.freebsd.arm64:
	sh scripts/build/build.sh $@

renovate-server.freebsd.all: \
	renovate-server.freebsd.amd64 \
	renovate-server.freebsd.arm64 \
	renovate-server.freebsd.armv7 \
	renovate-server.freebsd.x86 \
	renovate-server.freebsd.armv5 \
	renovate-server.freebsd.armv6

renovate-server.netbsd.amd64:
	sh scripts/build/build.sh $@

renovate-server.netbsd.x86:
	sh scripts/build/build.sh $@

renovate-server.netbsd.armv5:
	sh scripts/build/build.sh $@

renovate-server.netbsd.armv6:
	sh scripts/build/build.sh $@

renovate-server.netbsd.armv7:
	sh scripts/build/build.sh $@

renovate-server.netbsd.arm64:
	sh scripts/build/build.sh $@

renovate-server.netbsd.all: \
	renovate-server.netbsd.amd64 \
	renovate-server.netbsd.arm64 \
	renovate-server.netbsd.armv7 \
	renovate-server.netbsd.x86 \
	renovate-server.netbsd.armv5 \
	renovate-server.netbsd.armv6

renovate-server.openbsd.amd64:
	sh scripts/build/build.sh $@

renovate-server.openbsd.x86:
	sh scripts/build/build.sh $@

renovate-server.openbsd.armv5:
	sh scripts/build/build.sh $@

renovate-server.openbsd.armv6:
	sh scripts/build/build.sh $@

renovate-server.openbsd.armv7:
	sh scripts/build/build.sh $@

renovate-server.openbsd.arm64:
	sh scripts/build/build.sh $@

renovate-server.openbsd.all: \
	renovate-server.openbsd.amd64 \
	renovate-server.openbsd.arm64 \
	renovate-server.openbsd.armv7 \
	renovate-server.openbsd.x86 \
	renovate-server.openbsd.armv5 \
	renovate-server.openbsd.armv6

renovate-server.solaris.amd64:
	sh scripts/build/build.sh $@

renovate-server.aix.ppc64:
	sh scripts/build/build.sh $@

renovate-server.dragonfly.amd64:
	sh scripts/build/build.sh $@

renovate-server.plan9.amd64:
	sh scripts/build/build.sh $@

renovate-server.plan9.x86:
	sh scripts/build/build.sh $@

renovate-server.plan9.armv5:
	sh scripts/build/build.sh $@

renovate-server.plan9.armv6:
	sh scripts/build/build.sh $@

renovate-server.plan9.armv7:
	sh scripts/build/build.sh $@

renovate-server.plan9.all: \
	renovate-server.plan9.amd64 \
	renovate-server.plan9.armv7 \
	renovate-server.plan9.x86 \
	renovate-server.plan9.armv5 \
	renovate-server.plan9.armv6
