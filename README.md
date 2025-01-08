openscadgen is IN-DEVELOPMENT a tool for generating a set of .stl files from a single .scad file and a simple config file.

The goal of the tool is to ease the management and production of large numbers of stl files when working with one openSCAD file.

(early days and still in active development, please let me know if you encounter any issues)

# Why?

Seeing some frustrations in the community with the lack of 'remixabilty' when providing open source models, the openscad code does make remixes easier, but we can't expect everyone to be comfortable with (especially other peoples!) openscad code. This tools aims to make it easier to create a 'base' model, a list of useful instances of that model and then generate a set of stl files based on the 'base' model. A good analogy might be the difference between providing the source code and providing a set of pre-compiled binaries. The source code can be used to install the software, but the pre-compiled software for each platform will be more accessible to a wider audience.

Additionally, through 'version' string defined in the config file, we control the subfolder the export is made into, allowing for easy management of old versions.

The config file format is [toml](https://toml.io/en/)
```yaml
design:
  name: "screw-mounted-clip"
  description: "A parametric screw mounted clip"
  input_path: "./examples/screw-mounted-clip/parametricCommandStripBroomHook.scad"
  output_path: "./examples/screw-mounted-clip/export"
  version: v1.2
  instances:
    - name: "clip-for-hex-tool"
      description: "A clip sized for a hex tool with a large piece"
      params:
        handle_diameter: 5
        handle_offset: 10
    - name: "clip-short-and-strong"
      params:
        handle_diameter: 4
        handle_offset: 20
    - name: "clip-large"
      params:
        handle_diameter: 10
        handle_offset: 30

```

To run the tool, you need to have a config file and the OpenSCAD executable file.

To run, use: 
```
./openscadgen -c path/to/config.yml
```
(where config.yml is the openscadgen config file you want to generate for)


An [example config file](./examples/screw-mounted-clip/config.yml) is provided for the excellent "[Command Strip Parametric Broom Handle Clip](https://www.printables.com/model/516117-parametric-broom-handle-holder-openscad-command-st/related)" by [Brian Khuu](https://briankhuu.com/). 

# Installation

Currently the program is only available via the [github release page](https://github.com/kiwikid/openscadgen/releases)



# Development
```bash
# build
go build .

# run
./openscadgen -c ./examples/screw-mounted-clip/config.yml

# release   
# bump version in main.go
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"  
git push && git push --tags
```



## Troubleshooting


Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
exec: Failed to execute process: './openscadgen' the file could not be run by the operating system.
```

Solution:
```
Ensure you have downloaded the correct version of the tool from the releases page (i.e. darwin_arm64, linux_arm64, etc)
```


Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
fish: Job 1, './openscadgen -c examples/cup-h…' terminated by signal SIGKILL (Forced quit)
```

Solution:
```
Ensure you have allowed openscadgen to run (in Privacy & Security settings)
```





## TODO/Project Ideas
- [ ] (maybe) Add ability to generate instances via annotations in the scad file (i.e. remove need for config file)
```
// openscadgen: handle_offset: 10, 15, 25
handle_offset = 10
```
- [ ] Add ability to configure ranges of parameters in config file with auto-naming (i.e. handle_diameter: 5 to 10)
- [ ] (maybe) make stl builds parallel to speed up processing time 
- [ ] Warn when replacing existing stl export files
- [ ] Add clean-up option for old versions
- [ ] Add config file generation quickstart command (to create a config file from a scad file)
- [ ] (maybe) Add a watch mode to automatically re-run the tool when the scad file is changed

If you have any ideas/bugs/etc, please let me know and i'll try and fix them where possible. I do want to keep the goals of the project simple and specific as i believe it will result in the best tool for the job.