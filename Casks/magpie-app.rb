cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1741"
  sha256 arm:   "fdc244cc76f0198066d9394abc16851438f56265b1b5718a3304426729ff81c4",
         intel: "77d4d1d0468d5b4648da290782e6f2d9280882b67d3ed88db02d76c8d82c560b"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
