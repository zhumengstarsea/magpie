class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1554/magpie-cli-darwin-arm64"
      sha256 "66656cad32abe53d9df96c3fa4364789321bf75ab1194b7f8bcee29a755e984c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1554/magpie-cli-darwin-amd64"
      sha256 "bc4f2be9ea55edb96bf48111873d53828d930d022c2c4b97ed0a637ba252b189"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1554/magpie-cli-linux-arm64"
      sha256 "ca3ecff4b0f743918bbb970d0c52fed274f96bf25f845682df5a84655af0837f"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1554/magpie-cli-linux-amd64"
      sha256 "6ff9cdbd7d98df5922d2d132c2d6cd3ee479fd6ed0bc427575cb637d04c818e2"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
