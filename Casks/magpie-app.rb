cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1492"
  sha256 arm:   "ce219abc52ae0089fb114e2014d8ce4e39f504030f37ea706cae6bf087490571",
         intel: "6b7af71081067fc968b29928d4dbfeac298ce2d7090ac8a04d2714629f9a9cb7"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
