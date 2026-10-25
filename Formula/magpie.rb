class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1485/magpie-cli-darwin-arm64"
      sha256 "423f2cc8bce1a704662168b6fdb5038e9120213564fd9c07edd9ba322853bef5"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1485/magpie-cli-darwin-amd64"
      sha256 "ccb2a1fc3c51b739e6ec7ff927fbe8288567b79f98580ac2c23cc20f3434d555"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1485/magpie-cli-linux-arm64"
      sha256 "45d2b7861ffd388569fa57e5f293bbe11455492246a6224a09d78576cef5388c"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1485/magpie-cli-linux-amd64"
      sha256 "1fdcafbb1845c78b4acf236f7f50c5a71e1059e8fdf04c5bda3dfe988404a3ff"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
