class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1326/magpie-cli-darwin-arm64"
      sha256 "cc595f3f08deafe54173627f3db487422d5d77b49e6b16a4bfdce6bd24b3657c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1326/magpie-cli-darwin-amd64"
      sha256 "a083028062c6b37936bb10813e0dd7415376d0a0802147958c5da14e05db788b"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1326/magpie-cli-linux-arm64"
      sha256 "2e53a82296b5ebe862b49847400dc7b2be37db94f37d778cdfae6aeaeaa60c6a"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1326/magpie-cli-linux-amd64"
      sha256 "c41160057e4314ace2af0324ef5633ac3bc362b5ba59ead558efde20ea5ef45b"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
