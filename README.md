## An tool to generate Xephyr multiseat script

genseat can generate a script to multiseat on demond.

### Installation

 1. Install Xephyr, go, xinput, setxkbmap, sudo
 2. Plugin your extra keyboard and mouse.
 3. `go build` to generate `genseat` executable file

### Usage
 1. run `genseat > seat2.sh`
 2. edit seat2.sh before run. Modify following options depend on you.
  * modify DISPLAY(eg. :5 -> :10)
  * resolution (eg. 1024x768 -> 1280x1024)
  * terminal or window manager (eg. xterm -> awesome)
 3. Run it by `sh seat2.sh`
 4. Move the new seat to proper position by your first mouse.
