class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1298/magpie-cli-darwin-arm64"
      sha256 "fca5b32c6b3998605c26721b32b34d585c23c777acf9587854fdeef9d381dd84"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1298/magpie-cli-darwin-amd64"
      sha256 "0f7a6073ea28853be7ede42b4be6050d98722b2f927d1fbfece6f17f61e5c941"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1298/magpie-cli-linux-arm64"
      sha256 "1a5e92d0c891ec1719535e11709d4145679ad1f606a9d97e31c6cfcfd6b91faf"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1298/magpie-cli-linux-amd64"
      sha256 "7e192e425a0c3040a76c27f5a7560f98552e540ef129adb7d1918e0919b81733"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
