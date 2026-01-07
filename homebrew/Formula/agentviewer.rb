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
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"

      def install
        bin.install "agentviewer-darwin-arm64" => "agentviewer"
      end
    else
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"

      def install
        bin.install "agentviewer-darwin-amd64" => "agentviewer"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-linux-arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"

      def install
        bin.install "agentviewer-linux-arm64" => "agentviewer"
      end
    else
      url "https://github.com/pengelbrecht/agentviewer/releases/download/v#{version}/agentviewer-linux-amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"

      def install
        bin.install "agentviewer-linux-amd64" => "agentviewer"
      end
    end
  end

  test do
    assert_match "agentviewer version", shell_output("#{bin}/agentviewer --version")
  end
end
