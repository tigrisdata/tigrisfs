#!/usr/bin/python3
# Copyright 2025 Tigris Data, Inc.
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

import sys, os, threading, subprocess

def watchdog():
    print("fsync timed out")
    sys.exit(1)

path = subprocess.check_output("findmnt -f -o TARGET \""+sys.argv[1]+"\" | tail -n 1", shell = True).decode('utf-8').strip()
if not path:
    sys.exit(1)
timer = threading.Timer(60, watchdog)
timer.start()
fd = os.open(path, os.O_RDONLY)
os.fsync(fd)
os.close(fd)
subprocess.call("sudo /bin/umount "+path, shell = True)
timer.cancel()
