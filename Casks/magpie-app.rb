cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1322"
  sha256 arm:   "9fae03dd36832bae01e5012e2a295d9dcbc4f53d38e6298e6b987c76d956cd4f",
         intel: "d957a02529d5a2e13b35602947373c962364ec6655184b8afdffe190da44eb8e"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
