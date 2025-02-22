# car-holder-plus

A simple model that upgrades the center console cup holder from one small cup holder to two medium-large cup holders with a slot at either end for a phone
Tested against late 90s toyota models, you can adjust the cup holder size via these params in the openscad file:
Be-aware of the limitations of the material you are using, PLA can have a very limited lifespan as it will deform and warp in the sun/heat of the car (https://3dprinting.stackexchange.com/questions/6119/can-you-put-pla-parts-in-your-car-in-the-sun)

Project file includes:
- Cup Adapter without cut - an “All-in-on” print - uses more supports & time, stronger, with no assembly (~14h print time & 560g of PLA filament)
- Cut the Cup Adapter with a dowel cut - Minimise print/build time via splitting the cup holder to a separate part (~10h print time & 470g of PLA filament) - Parametric .scad file for customisation 


v1.1- Making the cup shorter and wider at the top (remove the need for tape)- Slightly narrower cup holder diameter at the brim and base- Reduce sharp edges (better align cup & phone holders in the center)

## Table of Contents
- [_A](#_a)
- [_B](#_b)

## _A
- **cup_holders_mode**: oneBigOneSmall
- **designFileName**: carHolderPlus
- **name**: rav4
- **in_car_cup_holder_height**: 70
- **in_car_cup_holder_top_diameter**: 74
- **in_car_cup_holder_bottom_diameter**: 66.5
- **center_offset**: 0

## _B
- **name**: rav4
- **in_car_cup_holder_height**: 70
- **in_car_cup_holder_top_diameter**: 74
- **in_car_cup_holder_bottom_diameter**: 66.5
- **center_offset**: 0
- **cup_holders_mode**: twoBig
- **designFileName**: carHolderPlus

## Additional Information
This README was generated by [openscadgen](https://github.com/KiwiKid/openscadgen) v1.1.7-ALPHA OpenSCAD version 2025.01.09. The free, local, open source openscad file generator.
