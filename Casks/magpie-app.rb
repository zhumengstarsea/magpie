cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1319"
  sha256 arm:   "ef0cc81d0833f83a8ef0755c43cd8270bac4bd864ba71541e0b8134b7ceaab2e",
         intel: "cd029056e198fc647913fd83761c6f91a7c9e487187672851b491148504e54f1"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
