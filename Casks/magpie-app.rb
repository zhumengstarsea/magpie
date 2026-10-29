cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.2040"
  sha256 arm:   "c36b58ac5c52f62b1aa758856435f9184f31cf249ed59fd10c3dc9a5a08017f6",
         intel: "892aabaa18a8519da2150a6dcd8d4ab7f8e0c6154a4721cdebf3b73dd56e2de6"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
