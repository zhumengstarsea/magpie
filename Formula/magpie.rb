class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1876/magpie-cli-darwin-arm64"
      sha256 "0650792d03b6fcdb0fc5240a62453dbf8c533ca890af401c1a51eacfd07b0cec"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1876/magpie-cli-darwin-amd64"
      sha256 "7cdf7fbf0d0f29e530be6bca9741149ef10cf7478f37dc3b4b05501ca95edcf3"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1876/magpie-cli-linux-arm64"
      sha256 "50a1e0ebe1b8bde5b3b8744af38affaef6d718c4009b31fe2ae10748b7ac5da7"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1876/magpie-cli-linux-amd64"
      sha256 "aa3fd839c6a23cf5a0530b9326c35c1793f2572d315529a37ec662956c96b825"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
