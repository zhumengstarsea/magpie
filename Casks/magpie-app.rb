cask "magpie-app" do
  arch arm: "arm64", intel: "amd64"

  version "0.1.1689"
  sha256 arm:   "51441574be3a9b0e54576d4edf77f0069b3787c8f223f324594029283ea33af9",
         intel: "b20aa1bf97fe5e1202b0049e883cd8437d9d0db6f96bd4c0807d46d4d323ac2b"

  url "https://github.com/zhumengstarsea/magpie/releases/download/v#{version}/magpie-darwin-#{arch}.zip"
  name "magpie"
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"

  auto_updates true
  depends_on :macos

  app "magpie.app"
end
