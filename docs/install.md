# Installation

We use Homebrew Cask for [distribution](https://github.com/TimeInc/homebrew-cask-tap/blob/master/Casks/ape-dev-rt.rb), see http://brew.sh/ and http://caskroom.io/ for installation instructions of both tools. Do the following when you have both installed:

```
brew tap TimeIncOSS/cask-tap git@github.com:TimeIncOSS/homebrew-cask-tap.git
brew cask install ape-dev-rt
```

To upgrade to a new version, run:

```
brew update && brew cask reinstall ape-dev-rt
```

**NOTE:**

This method of installation only works with access to the private cask repo
