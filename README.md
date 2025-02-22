openscadgen is IN-DEVELOPMENT a tool for generating a set of .stl files from a single .scad file and a simple config file.

The goal of the tool is to ease the management, production and distribution of large numbers of stl files when working with openSCAD.

(early days and still in active development, please let me know if you encounter any issues)

Additionally, through 'version' string defined in the config file, we control the subfolder the export is made into, allowing for easy management of old versions.

The config file format is [toml](https://toml.io/en/)

```toml
# These lines configure, where the config file is, how openscad will be run and where the output will be saved
[openscadgen]
# name of the design, will be used in the name of output files
name = "screw-mounted-clip"
# description of the design, will be used in the README.md file
description = "A parametric screw mounted clip"
# path to the openscad file that will be used to generate the design
input_path = "./examples/screw-mounted-clip/parametricCommandStripBroomHook.scad"
export_name_format = "clip-{handle_diameter}mm-wide-{handle_offset}mm-tall"
# path to the directory where the export (.stl files, README.md) will be saved
output_path = "./examples/screw-mounted-clip/export/"
# version of the design, the export will be saved in a subfolder with this version number
version = "v1.6"


# Dynamic Instances

# Each configuration below will result in a SET of separate .stl file being created with those parameters in the 'output_path' directory

[[openscadgen.dynamic_instances]]
# The 'name' field is a template string that will be used to generate the instance name (note the {param_name} syntax for value replacement)
# the params field configures which instances get created, this configures 50 
params = { handle_diameter = "5,7,8,10,15,20,25,30", handle_offset = "5,10,15,20,25,30" }


# Specific models
# Each instance below will result in a separate .stl file being created with those parameters in the 'output_path' directory
[[openscadgen.instances]]
name = "clip-for-large-ended-hex-tool"
description = "A clip sized for a hex tool with a large piece"
[openscadgen.instances.params]
handle_diameter = 7
handle_offset = 15

```

This is a more complex example, where we have a main scad file, and a separate scad file to provide a utility sizer.

```toml
# These lines configure, where the config file is, how openscad will be run and where the output will be saved
[openscadgen]
name = "car-holder-plus"
description = """\
  A simple model that upgrades the center console cup holder from one small cup holder to two medium-large cup holders with a slot at either end for a phone
\n
Tested against late 90s toyota models, you can adjust the cup holder size via these params in the openscad file:
\n
Be-aware of the limitations of the material you are using, PLA can have a very limited lifespan as it will deform and warp in the sun/heat of the car (https://3dprinting.stackexchange.com/questions/6119/can-you-put-pla-parts-in-your-car-in-the-sun)

\n\n
Project file includes:
\n
- Cup Adapter without cut - an “All-in-on” print - uses more supports & time, stronger, with no assembly (~14h print time & 560g of PLA filament)
\n
- Cut the Cup Adapter with a dowel cut - Minimise print/build time via splitting the cup holder to a separate part (~10h print time & 470g of PLA filament)
 - Parametric .scad file for customisation 
\n\n\n
v1.1
- Making the cup shorter and wider at the top (remove the need for tape)
- Slightly narrower cup holder diameter at the brim and base
- Reduce sharp edges (better align cup & phone holders in the center)
"""
export_name_format = "{cup_holders_mode}/{designFileName}-{name}"

output_path = "./examples/cup-holder-plus/export/"
version = "v1.3"

[[openscadgen.input_paths]]
path = "./examples/cup-holder-plus/carHolderPlus.scad"

[[openscadgen.input_paths]]
path = "./examples/cup-holder-plus/carHolderPlus-cup-holder-sizer.scad"
filter_params = "cup_holders_mode"
export_name_format = "cup-holder-sizer/{designFileName}-{name}"

[[openscadgen.dynamic_instances]]
params = { name= "rav4", in_car_cup_holder_height = "60", in_car_cup_holder_top_diameter = "74", in_car_cup_holder_bottom_diameter = "66.5", cup_holders_mode = "twoLargerHolders,oneLargerHolderOneSmallerHolder" }

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


## Troubleshooting


Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
exec: Failed to execute process: './openscadgen' the file could not be run by the operating system.
```
Ensure you have downloaded the correct version of the tool from the releases page (i.e. darwin_arm64, linux_arm64, etc)

Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
fish: Job 1, './openscadgen -c examples/cup-h…' terminated by signal SIGKILL (Forced quit)
```
Ensure you have allowed openscadgen to run (in Privacy & Security settings)




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


# Why?

Seeing some frustrations in the community with the lack of 'remixabilty' when providing open source models, the openscad code does make remixes easier, but we can't expect everyone to be comfortable with (especially other peoples!) openscad code. This tools aims to make it easier to create a 'base' model, a list of useful instances of that model and then generate a set of stl files based on the 'base' model. A good analogy might be the difference between providing the source code and providing a set of pre-compiled binaries. The source code can be used to install the software, but the pre-compiled software for each platform will be more accessible to a wider audience.





## TODO/Project Ideas
- [ ] Allow for multi-part builds in the same config file (i.e. multiple scad files, generated with the same input parameters)
- [ ] Allow for concurrent openscad runs
- [ ] Directory generation (i.e. dynamically find and generate all the instance configs in a directory)
- [ ] (maybe) Add ability to generate instances via annotations in the scad file (i.e. remove need for config file). Something like:
```
// openscadgen: handle_offset: 10, 15, 25
handle_offset = 10
```
- [ ] Allow for building a whole directory of scad files (i.e. generate all the instances in a directory)
- [ ] Split export folder into 'base' and 'has_part_letter' folders
- [ ] Allow for setting of part id in the static instance config
- [ ] (maybe) make stl builds parallel to speed up processing time 
- [ ] Add clean-up option for old versions
- [ ] Add config file generation quickstart command (to initialise a openscadgen config file from a scad file)
- [ ] (maybe) Add configurable watch mode to automatically re-run the tool when the scad file is changed
- [ ] Tidy/Improve logging and log handling
- [ ] Tools to ease handling of common openscad external libraries (i.e. BOSL2 etc)
- [-] Warn when replacing existing stl export files
- [-] Add ability to configure ranges of parameters in config file with auto-naming (i.e. i want models for each handle_diameter from 5 to 10, with )

If you have any ideas/bugs/etc, please let me know and i'll try and fix them where possible. I do want to keep the goals of the project simple and specific.