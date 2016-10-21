// This file is a part of the IncludeOS unikernel - www.includeos.org
//
// Copyright 2015 Oslo and Akershus University College of Applied Sciences
// and Alfred Bratterud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <os>
#include <net/inet4>

constexpr int port {8080};

void Service::start(const std::string&) {
  using namespace std::string_literals;

  printf("here we go!\n");
  printf("OS::ready_ is %d\n", OS::ready_);

  auto& server = net::Inet4::stack().tcp().bind(port);
  server.on_connect([] (auto conn) {
    conn->on_read(1024, [conn] (auto buf, size_t n) {
      auto response {"My first unikernel!\n"s};
      conn->write(response);
      conn->close();
    });
  });
}
