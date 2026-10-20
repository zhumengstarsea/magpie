class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1302/magpie-cli-darwin-arm64"
      sha256 "1b79a6397544df522bcc35548b68d3f70e87786b15bf674cf3ec181d42b3de09"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1302/magpie-cli-darwin-amd64"
      sha256 "0f205c2d210cef9bc6d3831158bb1c52ccb3ffc4ac21ea0e88c1ae795b686a0e"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1302/magpie-cli-linux-arm64"
      sha256 "f3566944cdd344bb872a8bff794af14d4a3b23fc1ba6a9be1d30a3e3852df230"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1302/magpie-cli-linux-amd64"
      sha256 "48aad15a17543ec928576273c3147ff377936b92ee34a308435cc4c83537c08e"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
