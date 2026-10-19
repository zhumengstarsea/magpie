class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-darwin-arm64"
      sha256 "c7943c85a2ca4b6abf46796d2b57b39cffca803af471f4a0cac30022de94b659"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-darwin-amd64"
      sha256 "ae557ccd8fc2983923d6324f3e9164e605cfc5372ce86ddcde9f869df1a73187"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-linux-arm64"
      sha256 "8f794b3c6854f5a946fcdaa98351603b871c715aa39cf3784e0ff81d909a6ba6"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1294/magpie-cli-linux-amd64"
      sha256 "55bcd3a78a5bd9f5fb3df4c50019c04efa074479e84bfc7c7e24b6e2e6496663"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
