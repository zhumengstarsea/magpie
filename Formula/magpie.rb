class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1316/magpie-cli-darwin-arm64"
      sha256 "87726f8ff2ca547031f1eeea833d3d5bee6ceddea4b081fca606e2bd62ef61cf"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1316/magpie-cli-darwin-amd64"
      sha256 "abf94a7e22469e771608b906f28511181170cb64d2a823df03307ecc257d916e"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1316/magpie-cli-linux-arm64"
      sha256 "ab52a22f3edf5bba18398484c7b825292a0c6ca144cd976fe054663322922f1e"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1316/magpie-cli-linux-amd64"
      sha256 "eab103dbe0589d8c4acd08523172c342dc214e3a885363ce5bc7aea157147050"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
