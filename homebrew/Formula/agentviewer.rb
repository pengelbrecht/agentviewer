# typed: false
# frozen_string_literal: true

class Agentviewer < Formula
  desc "CLI tool for AI agents to display rich content in a browser"
  homepage "https://github.com/pengelbrecht/agentviewer"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-darwin-arm64"
      sha256 "caeabe94f97d4ea26095cee510ca176c7607411aa03cfc0575e7fecca4e71e9d"

      def install
        bin.install "agentviewer-darwin-arm64" => "agentviewer"
      end
    else
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-darwin-amd64"
      sha256 "525a990c6ca4c0803684244916f6c54c3c21e4e8524c272a224ee6c4e8d74e3e"

      def install
        bin.install "agentviewer-darwin-amd64" => "agentviewer"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-linux-arm64"
      sha256 "50208997e697baf98c49fc5dcfe22b3713d349e42af676aed54e96d3fb18cf1a"

      def install
        bin.install "agentviewer-linux-arm64" => "agentviewer"
      end
    else
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-linux-amd64"
      sha256 "fa016b07e0e5dea80a44369c76904a8e3b8acc19b17f6bae5be832523b7f9a5d"

      def install
        bin.install "agentviewer-linux-amd64" => "agentviewer"
      end
    end
  end

  test do
    assert_match "agentviewer version", shell_output("#{bin}/agentviewer --version")
  end
end
