cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1788"
  sha256 arm:   "849a47ee6a2dcab412311b5ea99e375a171c0aa6df98dd3997764b0982050672",
         intel: "e49ab1449771e3ec2ffa79b1119a24b5c679b0f4c117e631ea81439eeb021af9"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
