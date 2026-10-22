class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1319/magpie-cli-darwin-arm64"
      sha256 "8de1534b00d77c46f12cc778a83219a830b3e7407c77f273afdc26a6d08dff07"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1319/magpie-cli-darwin-amd64"
      sha256 "85f1031564f703d345112dabcf5500e3ad82c7f0aadeaebda2913f5f55bdc7e9"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1319/magpie-cli-linux-arm64"
      sha256 "c31d96b41de56da57ac33094b83e86040b201f867cc8cad9f7caf7c032564dbc"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1319/magpie-cli-linux-amd64"
      sha256 "33bd902a48558cf05c0b028bc3db6a39bab7410ebc11f89b43beaa9d8ec00a1f"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
