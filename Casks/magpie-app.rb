cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1982"
  sha256 arm:   "4e3560b51c16e40fce5c22fb7d9051dfd78338be4bacd34e84072b9e07d3c204",
         intel: "cfe8fc38f05b445aea7982434c9359660ac9f144ca988668eb0ab73c322dd1b3"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
