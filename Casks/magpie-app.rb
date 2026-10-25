cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1554"
  sha256 arm:   "0427c246821c9e6a61d0f0ca96eb72311f850e91035697322e3cd6f46e3bf388",
         intel: "730fb4b979c5ff3ee6b1080b4d0333b148db4ec135077a4df72c1fa5eec4f40d"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
