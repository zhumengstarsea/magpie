cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1326"
  sha256 arm:   "a1cb22cd9128f455ba1d600dde12eee1ce38471cc5809596a171823f72cddeb4",
         intel: "37c37a3d31d901d5900035e4fb9ba364b02a9168bcccac12c71cce9ab42d9ace"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
