class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-darwin-arm64"
      sha256 "000194ada29bb0871b47a40f81a12fc81bb11f60084b41772e7cda7759341c5c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-darwin-amd64"
      sha256 "7b5ff376944c48d41fd5b98cfeffd909904179cf2c362c74d0c5f583f3f2c4a2"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-linux-arm64"
      sha256 "1b663f0cf50f1d54e177928c6600ee920f1957932a50e384d7631efab5085bdf"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-linux-amd64"
      sha256 "e9217d92cf0a68c2627fb4a5fc2d180b207898219390dbb800fbae9f1abed1db"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
