class Magpie < Formula
  desc "Pick the model each AI coding agent on your machine uses"
  homepage "https://github.com/zhumengstarsea/magpie"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1587/magpie-cli-darwin-arm64"
      sha256 "b6e7c3f8ef5ee953de2b6c23d2e611fb234ef4eb70c46b6a75fbf42ad801dbdd"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1587/magpie-cli-darwin-amd64"
      sha256 "eda03e1ffc4bc8af19bcb1daf78f1a9f930161611cd15b6d1c514d724c20c90e"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1587/magpie-cli-linux-arm64"
      sha256 "2749e2335cc7fc8c234e9073d3d90edd3793e37513a819bb5f13ab3fcef60306"
    end
    on_intel do
      url "https://github.com/zhumengstarsea/magpie/releases/download/v0.1.1587/magpie-cli-linux-amd64"
      sha256 "018c6aa8e7ce48b17d8325b9b4fab8063987c74613339ca746a42717049d62b0"
    end
  end

  def install
    bin.install Dir["magpie-cli-*"].first => "magpie"
  end

  test do
    assert_match "magpie #{version}", shell_output("#{bin}/magpie --version")
  end
end
