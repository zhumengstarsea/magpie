cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1559"
  sha256 arm:   "c81c5d994f5e68ef07dca58da80b6b31fd55dc00b6a158217497d9f5bc5bde2d",
         intel: "cba1e309dde0cc1931efeb8ea51dd445c63c987db963e52456d157fb9fadf8b3"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
