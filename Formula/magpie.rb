class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1492/magpie-cli-darwin-arm64"
      sha256 "a42f807e51c3d0127629a3a7f842a7b33a39617c9298da2ab928f859f758519b"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1492/magpie-cli-darwin-amd64"
      sha256 "88e1f4f3fd2e864f294fe82d93b7b9590204db71bd2d0ff3966c31573751ec86"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1492/magpie-cli-linux-arm64"
      sha256 "2d04a35d7058b5bac0e3a17a3b76f52eaf10ebf0562a95b4fd0a29622b33305c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1492/magpie-cli-linux-amd64"
      sha256 "6f50e11edc0dd61be39d20b532911e5143cc88be5794f4b4d8460428c8aa1caa"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
