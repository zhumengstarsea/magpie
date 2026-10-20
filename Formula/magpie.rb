class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1296/magpie-cli-darwin-arm64"
      sha256 "340f6d0001346100de1c1e548e7e41c3bf10d0f3dd8d2be4288a472f1e694da4"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1296/magpie-cli-darwin-amd64"
      sha256 "8e9da4c0753ab05ccbe5819e76e74ed49c582dd910e5342655b309fc96d8eb25"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1296/magpie-cli-linux-arm64"
      sha256 "97e1888c2cba1121f40aa9cf892c288398ffc530d2d146b442a5e860f1a65d1c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1296/magpie-cli-linux-amd64"
      sha256 "b65d3bdb633728796bfb392248a9fa83d6d9d79933e439b70235fd7ef8de1e3b"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
