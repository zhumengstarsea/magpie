cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1616"
  sha256 arm:   "f62bd66d6a87f134b524cc2fbde641345266cacf8604aba155dfd682c6a69e29",
         intel: "0eb2b3f3995300dee085e023067acc9586f8688c4f10e9ab8d8e0e88f18086bf"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
