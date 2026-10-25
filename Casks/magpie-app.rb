cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1488"
  sha256 arm:   "c7362ad0b62ad1a755286faa6f07fdf360df89c27c8930901b1fd451c454e8bc",
         intel: "534094b458e17939161a80bc8e3820c48ea8af171cf61dee5fafc54b4930c6be"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
