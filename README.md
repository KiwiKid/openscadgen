openscadgen is IN-DEVELOPMENT a tool for generating a set of .stl files from a single .scad file and a simple config file.

(early days, please let me know if you encounter any issues)

The goal of the tool is to ease the management and production of large numbers of stl files when working with one openSCAD file.

Seeing some frustrations in the community with the lack of 'remixabilty' when providing open source models, the openscad code does make remixes easier, but we can't expect everyone to be comfortable with (especially other peoples!) openscad code. This tools aims to make it easier to create a 'base' model, a list of useful instances of that model and then generate a set of stl files based on the 'base' model.



openscadgen uses the original .scad file and a specified config file set to generate a set of stl files based on configured 'instances'

The config file format is:
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



## TODO/Project Ideas
- [ ] Add ability to configure ranges of parameters (i.e. handle_diameter: 5 to 10)
- [ ] (maybe) make stl builds parallel to speed up processing time 
- [ ] Warn when replacing existing stl export files
- [ ] Add clean-up option for old versions
- [ ] Add config file generation quickstart command (to create a config file from a scad file)
- [ ] (maybe) Add a watch mode to automatically re-run the tool when the scad file is changed

If you have any ideas/bugs/etc, please let me know and i'll try and fix them where possible. I do want to keep the goals of the project simple and specific as i believe it will result in the best tool for the job. s