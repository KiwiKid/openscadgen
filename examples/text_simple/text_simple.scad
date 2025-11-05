include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


content="Greg";
text_size=40;
h=50;
font="Apple Chancery:style=Chancery";//"Baloo 2";
spacing=.7;
center=false;
spin=0;
orient=[0,0,-1];
hasBase = true;
baseHeight= 30;
rot= 30;

textOffset=20;
rounding=10;

module text_simple(
    content,
    text_size,
    h,
    font,
    spacing,
    center,
    spin,
    orient,
    hasBase,
baseHeight,
rounding
){
    if(hasBase){
    difference(){
    
        #cuboid([500,text_size*1.4,h*0.9]);
        text3d(
            content,
            h=h,
            size=text_size,
            font=font,
            spacing=spacing,
            center=true,
            spin=spin,
            orient=orient,
            rounding=rounding
        );
        }
    } else{
    difference(){
    union(){
    
    move([0,0,-textOffset])
    rotate([rot,0,0])
    text3d(
        content,
        h=h,
        size=text_size,
        font=font,
        spacing=spacing,
        center=center,
        spin=spin,
        orient=orient,
                    rounding=rounding

    );
    
    
    move([-120,0,textOffset])
    rotate([rot,180,0])
    text3d(
        content,
        h=h,
        size=text_size,
        font=font,
        spacing=spacing,
        center=center,
        spin=spin,
        orient=orient
    );
    }
    
    fwd(40)
    #cuboid([300,100,300]);
    }
    }
    
}

text_simple(
    content=str(content),
    text_size=text_size,
    h=h,
    font=font,
    spacing=spacing,
    center=center,
    spin=spin,
    orient=orient,
    hasBase=hasBase,
    baseHeight=baseHeight,
    rounding=baseHeight
);
        