openscadgen is a simple tool for generating a set of .stl files (or other openscad export formats) from a .scad file(s) and a simple human-readable toml config file. A html report is generated containing the details of the generated instances
 
The goal of the tool is to ease the development, management, production and distribution of large numbers designs or design options based on single openscad file.

A simpler alterative to programmatic wrappers like Makefiles, AnchorSCAD, LuaCAD, PythonSCAD - designed provides a more accessible/structured approach

https://github.com/user-attachments/assets/61d605fa-3b5a-43c0-98d2-71357f96fc32


> [!WARNING]
> Early days and still in active development, please let me know if you encounter any issues or have ideas for improvements.


## Basic Configuration

The simplest configuration uses:
-  one input scad file
and creates multiple .stl files based on a combination of params against each [[openscadgen.instances]] :

```toml
[openscadgen]
name = "simple-clip"
description = "A basic parametric clip"
input_path = "./clip.scad" # input_path is relative to the config.toml file
export_name_format = "clip-{width}mm-{height}mm"
version = "v1.0"

[[openscadgen.instances]]
params = { width = "10,20,30", height = "5,10,15" }
```
When run via:
```
./openscadgen -c ./config.toml
```
openscadgen will:
- Use a single input file (`clip.scad`)
- Generate all combinations of width and height
- Export STLs at the export_name_format with the format `clip-10mm-5mm.stl`, `clip-10mm-10mm.stl`, etc.

## Multiple Input Files

When you need to generate multiple designs using the same parameters - multiple [[openscadgen.input_paths]] can be used

```toml
[openscadgen]
name = "cup-holder"
description = "A cup holder with multiple components"
export_name_format = "{designFileName}-{diameter}mm"
version = "v1.0"

[[openscadgen.input_paths]]
path = "./cup-holder-main.scad"

[[openscadgen.input_paths]]
path = "./cup-holder-adapter.scad"

[[openscadgen.instances]]
params = { diameter = "70,80,90", style = "minimal,decorative" }
```

Both designs are rendered using the same diameter parameters, with the adapter having its own export path format  - the filename of the source .scad files are replaced as {designFileName} in the `export_name_format`

## Adding Image Previews

To generate preview images of your STL files:

```toml
[openscadgen]
name = "phone-stand"
description = "A parametric phone stand"
input_path = "./phone-stand.scad"
export_name_format = "stand-{angle}deg"
version = "v1.0"

[[openscadgen.instances]]
params = { angle = "45,60,75" }

[[openscadgen.export_images]]
name = "top"

[[openscadgen.export_images]]
name = "front"

[[openscadgen.export_images]]
name = "side"
coords = "0,0,0,90,0,0,600"
```

This adds preview images from multiple angles for each generated STL.

## Using Parameter Sets

For more complex designs, parameter sets help organize related parameters:

```toml
[openscadgen]
name = "complex-bracket"
description = "A bracket with multiple configurations"
input_path = "./bracket.scad"
export_name_format = "bracket-{size}-{style}"
version = "v1.0"

[[openscadgen.param_sets]]
name = "small"
params = { width = "20", height = "30", thickness = "3" }

[[openscadgen.param_sets]]
name = "large"
params = { width = "40", height = "60", thickness = "5" }

[[openscadgen.param_sets]]
name = "minimal"
params = { style = "minimal", edge_radius = "2" }

[[openscadgen.param_sets]]
name = "decorative"
params = { style = "decorative", edge_radius = "5", pattern = "diamond" }

[[openscadgen.instances]]
name = "small-minimal"
param_sets = "small,minimal"

[[openscadgen.instances]]
name = "large-decorative"
param_sets = "large,decorative"
```

This configuration:
- Defines reusable parameter sets
- Combines parameter sets to create different configurations
- Makes it easier to manage complex parameter combinations

The config file format uses [toml](https://toml.io/en/) and some examples are include below and more are in the 'examples' directory

To run the tool, you need to have a config file and OpenSCAD executable file available on PATH.

To run, use: 
```
./openscadgen -c path/to/config.yml
```
(where config.yml is the openscadgen config file you want to generate for)


An [example config file](./examples/screw-mounted-clip/config.yml) is provided for the excellent "[Command Strip Parametric Broom Handle Clip](https://www.printables.com/model/516117-parametric-broom-handle-holder-openscad-command-st/related)" by [Brian Khuu](https://briankhuu.com/). 

# Installation 
Available via the [github release page](https://github.com/kiwikid/openscadgen/releases)




# Helpful options
```flag
-c string
        Path to config file 
         the absolute or relative path to the .toml config file
  -coe
        Continue on error
  -config string
        Path to config file 
         the absolute or relative path to the .toml config file
  -continue-on-error
        Alias for -co
  -custom-openscad-command string
        Custom OpenSCAD command to use
  -d    Alias for -debug
  -debug
        debug mode, more output
  -el
        Alias for -include-export-log-file
  -fi
        Set build info in file attributes (default true) (default true)
  -fn int
        Override the default fn value (default none)
  -h    Alias for -man
  -hq
        Set high quality (fn = 200)
  -i string
        Alias for -init
  -ie string
        Alias for -init
  -include-export-log-file
        Include the export log in the README.md file
  -init string
        Initialize a new project at the current directory with the given name
  -init-extended string
        Initialize a new project at the current directory with the given name - with bosl2 and renderSlicing support
  -lq
        Set low quality (fn = 20)
  -m    Alias for -man
  -man
        Display help message
  -n int
        Maximum number of instances to process
  -no-processing
        'dry-run' mode - will check config and provide instances that will be processed, but not do any processing
  -np
        Alias for -no-processing
  -overwrite
        Alias for -ow
  -ow
        Overrwite existing files
  -pid
        Include optional_part_id_letter variable in the call the openscad
  -q    Alias for -quiet
  -quiet
        quiet mode, no log output
  -r string
        Alias for -regex
  -regex string
        Regex pattern to filter instances by name
  -sd
        Alias for -skip-readme
  -skip-docs
        Skip generating a README.md file
  -skip-render
        Dont run a render before export
  -sr
        Alias for -skip-render
  -v    Alias for -version
  -version
        just output the openscadgen and openscad version number
```


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


# Config file format




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
# version of the design, the export will be saved in a subfolder with this version number
version = "v1.6"


# Instances

# Each configuration below will result in a SET of separate .stl file being created with those parameters in the 'output_path' directory

[[openscadgen.instances]]
# The 'name' field is a template string that will be used to generate the instance name (note the {param_name} syntax for value replacement)
# the params field configures which instances get created, this configures 50 
params = { handle_diameter = "5,7,8,10,15,20,25,30", handle_offset = "5,10,15,20,25,30" }


# Specific models
# Each instance below will result in a separate .stl file being created with those parameters in the 'output_path' directory
[[openscadgen.instances]]
name = "clip-for-large-ended-hex-tool"
description = "A clip sized for a hex tool with a large piece"
params = {  handle_diameter = "7", handle_offset="15" }
```


This is a more complex example, where we have two "Input Files" file, with params shared between each

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
- Cup Adapter without cut - an "All-in-on" print - uses more supports & time, stronger, with no assembly (~14h print time & 560g of PLA filament)
\n
- Cut the Cup Adapter with a dowel cut - Minimise print/build time via splitting the cup holder to a separate part (~10h print time & 470g of PLA filament)
 - Parametric .scad file for customisation 
\n\n\n
v1.1
- Making the cup shorter and wider at the top (remove the need for tape)
- Slightly narrower cup holder diameter at the brim and base
- Reduce sharp edges (better align cup & phone holders in the center)
"""
export_name_format = "{cup_holders_mode}/{designFileName}-{instanceName}"
version = "v1.3"

[[openscadgen.input_paths]]
path = "./carHolderPlus.scad"

[[openscadgen.input_paths]]
path = "./carHolderPlus-cup-holder-sizer.scad"
ignore_param_when_processing = "cup_holders_mode"

[[openscadgen.instances]]
params = { name= "rav4", in_car_cup_holder_height = "60", in_car_cup_holder_top_diameter = "74", in_car_cup_holder_bottom_diameter = "66.5", cup_holders_mode = "twoLargerHolders,oneLargerHolderOneSmallerHolder" }

```


# Development
```bash
# build
go build .

# test
go test -v ./tests

# run
./openscadgen -c ./examples/screw-mounted-clip/config.yml

# test



# release   
# bump version in main.go
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"  
git push && git push --tags
```

## TODO/Project Ideas
- [ ] Better validation/messaging around config issues
- [ ] Allow for concurrent openscad runs
- [ ] Directory generation (i.e. dynamically find and generate all the instance configs in a directory)
- [ ] (maybe) Add ability to generate instances via annotations in the scad file (i.e. remove need for config file). Something like:
```
// openscadgen: handle_offset: 10, 15, 25
handle_offset = 10
```
- [ ] Use new Go Tool include features to include openscad directly
- [ ] Allow for building a whole directory of scad files (i.e. generate all the instances in a directory)
- [ ] Split export folder into 'base' and 'has_part_letter' folders + refine part letter generation
- [ ] Allow for setting of part id in the static instance config
- [ ] Allow for exporting a set of images
- [ ] (maybe) Configure option to automatically create a horizontal and/or vertical cross-sectional slice of the part
- [ ] Add clean-up option for old versions
- [ ] (maybe) Add configurable watch mode to automatically re-run the tool when the scad file is changed
- [ ] Tidy/Improve logging and log handling
- [ ] Tools to ease handling of common openscad external libraries (i.e. BOSL2 etc)
- [-] Better handling of params with Spaces in them (i.e. name= "My Name" -> "My-Name.stl")
- [-] Better handling of params with Spaces in them (i.e. name= "My Name" -> "My-Name.stl")
- [-] Add a HTML export file + image generation
- [-] Warn when replacing existing stl export files
- [-] Add ability to configure ranges of parameters in config file with auto-naming (i.e. i want models for each handle_diameter from 5 to 10, with )
- [-] Allow for multi-part builds in the same config file (i.e. multiple scad files, generated with the same input parameters)
- [-] Add config file generation quickstart command (to initialize a openscadgen config file from a scad file) (`openscadgen i new-project-name`)


If you have any ideas/bugs/etc, please let me know and i'll try and fix them where possible. I do want to keep the goals of the project simple and specific
