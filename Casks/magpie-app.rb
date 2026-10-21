cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1316"
  sha256 arm:   "1c154b21e94f48980e928d90599b6e75aa9e7544f67f292f920c1b26c2931f9b",
         intel: "fac352b4be520414129d491a8961a45213d3ae3ec88a99cd79766aed70a21eab"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
