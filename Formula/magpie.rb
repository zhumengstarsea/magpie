class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.2040/magpie-cli-darwin-arm64"
      sha256 "62464fcca613f1ea238bc2c33025203fbed520903dc0207c9b95db555874c381"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.2040/magpie-cli-darwin-amd64"
      sha256 "2e715d95c5c1f9077b6122a7788fbeea2a45c2d6d41fed5f6c49734705e0a829"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.2040/magpie-cli-linux-arm64"
      sha256 "28b573b0dfa55996922ff00b6a177a3fc7acfd755956972b5ddff2572da6c8d1"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.2040/magpie-cli-linux-amd64"
      sha256 "15851a740b01ea176f2f8c06b202d3977ad45a1c07574d41042c4e31ff0fccd5"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
