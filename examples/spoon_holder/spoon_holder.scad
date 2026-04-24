
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

holderLength = 60;

holderWidth = 20;

spoonSplitWidthRatio = 0.6;

holderHeight = 10.65;
holderRoundingReduction = 0.93;
spitHeightRatio = 1.3;

spoonSplitThickness = 1.4;
spoonSplitSize = [spoonSplitThickness,holderWidth*spoonSplitWidthRatio,holderHeight*spitHeightRatio];
spoonSplitTranslate = [-holderLength*0.4,0,0.2];
spoonSplitRotate = 15;
module spoon_holder(){
   difference(){
   
	cuboid([holderLength,holderWidth, holderHeight], rounding=holderHeight*holderRoundingReduction, edges=TOP);
    translate(spoonSplitTranslate)
    yrot(spoonSplitRotate){
        cuboid(spoonSplitSize, rounding=-8, edges=TOP);
    }

    }
}


spoon_holder();
