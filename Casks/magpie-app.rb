cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1306"
  sha256 arm:   "36e786bc87f86d03961e8931085c7a4143a62bbd2a873e9235ca163914420ba3",
         intel: "7c18e638705c3cfcb5d94f2a2dced7d4d5e732ddc23056c9d00ec248e73e1dce"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
