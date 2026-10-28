cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1876"
  sha256 arm:   "f9c06105a9f0c1be1b4dd005cd1c940e70b23151dafc43882bb4dabde265a2a6",
         intel: "2d7fa574512f7ca2432e06a988c505bb3304f3476585a3dda63452d343ae7bee"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
