class ViMongo < Formula
  desc "Terminal User Interface for MongoDB"
  homepage "https://github.com/kopecmaciej/vi-mongo"
  version "v0.1.29"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/kopecmaciej/vi-mongo/releases/download/v0.1.29/vi-mongo_Darwin_arm64.tar.gz"
      sha256 "39579534da44bd67f52509dfa1ec9132d7774a95ab3303c81260bc160696ab90"
    else
      url "https://github.com/kopecmaciej/vi-mongo/releases/download/v0.1.29/vi-mongo_Darwin_x86_64.tar.gz"
      sha256 "fda573e5d183b586cdded197a84385455fada520d4a49b2b861cf8978a1b36a3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/kopecmaciej/vi-mongo/releases/download/v0.1.29/vi-mongo_Linux_arm64.tar.gz"
      sha256 "39579534da44bd67f52509dfa1ec9132d7774a95ab3303c81260bc160696ab90"
    else
      url "https://github.com/kopecmaciej/vi-mongo/releases/download/v0.1.29/vi-mongo_Linux_x86_64.tar.gz"
      sha256 "fda573e5d183b586cdded197a84385455fada520d4a49b2b861cf8978a1b36a3"
    end
  end

  def install
    bin.install "vi-mongo"
  end

  test do
    system "#{bin}/vi-mongo", "--version"
  end
end 
